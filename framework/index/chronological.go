package index

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/retro-framework/go-retro/framework/ctxkey"
	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/ref"
	"github.com/retro-framework/go-retro/framework/retro"
)

// Chronological returns a streaming index
// for the given branch name. which is wrapped up
// in a channel-like API. Cancel the given context
// to indicate that the Chronological should
// flush and close it's channel. TODO: is this right?
func Chronological(ctx context.Context, odb object.DB, refdb ref.DB) <-chan retro.Hash {

	var out = make(chan retro.Hash)

	var ci = &checkpointChronologiclIndex{
		objdb:       odb,
		refdb:       refdb,
		toc:         tocCPs{},
		cp:          make(chan tocCP),
		rootReached: make(chan struct{}),
	}

	go ci.run(ctx, out)

	// Bootstrapping, index will build from
	// here (retrieveCheckpoint recurses)
	ref, err := refdb.Retrieve(ctxkey.Ref(ctx))
	if err != nil {
		panic("index: error retrieving ref")
	}
	ci.wg.Add(1)
	ci.retrieveCheckpoint(ref)

	// Fo not return until we have completed the index.
	// TODO: this could be done more elegantly I think, so that
	// consumers block on reading the channel, not on calling this
	// constructor
	ci.wg.Wait()

	return out
}

type checkpointChronologiclIndex struct {
	// storage, object and our working slice
	objdb object.DB
	refdb ref.DB

	toc tocCPs

	// mechanics for tracking whether we're complete
	// we receive all the retrieved checkpoints
	// so we can track how many to sign off on.
	// we increment the wg for each entry in the parents
	// hash that we see, and decrement it every time we process
	// one, guarded by a for{select{}} to avoid memory races
	sync.Mutex
	wg     sync.WaitGroup
	cp     chan tocCP
	cursor int

	// mechanisms for knowing if we reached the root, yet
	// the retrieveCheckpoint func will send on rootReached
	// when it encouters a checkpoint with no parents. The
	// run() func will consume this and set startConsumer
	// to indicate that consumers may now start consuming
	// chronologically.
	rootReached    chan struct{}
	consumerActive bool
}

func (ci *checkpointChronologiclIndex) run(ctx context.Context, out chan<- retro.Hash) {
	var (
		o    chan<- retro.Hash
		next retro.Hash
		once sync.Once
	)

	for {
		select {
		case o <- next:

			// make sure we're not going to return to this branch
			o = nil
			next = nil

			// advance the cursor
			ci.cursor++

			// check if we are still in-bounds, and enable this branch
			if ci.cursor < len(ci.toc) {
				o = out
				next = ci.toc[ci.cursor].h
			}

		// This branch receives new checkpoints from the goroutine
		// which is exploring the index. It must mark wg.Done() to
		// keep the wg accurate. In the case that we are active for
		// consumers, we also check if we are in bounds, and enable
		// the producer branch by setting next/o to the correct
		// values
		case t := <-ci.cp:
			ci.toc = ci.toc.Insert(t)
			ci.wg.Done()

			if ci.cursor < len(ci.toc) && ci.consumerActive {
				o = out
				next = ci.toc[ci.cursor].h
			}
		case <-ci.rootReached:
			once.Do(func() {
				o = out
				ci.consumerActive = true
				if len(ci.toc) > 0 {
					next = ci.toc[0].h
				}
			})
		case <-ctx.Done():
			ci.rootReached = nil
			o = nil
			ci.cp = nil
			return
		}
	}
}

func (ci *checkpointChronologiclIndex) retrieveCheckpoint(h retro.Hash) {

	var (
		jp         *packing.JSONPacker
		checkpoint packing.Checkpoint
		err        error
	)

	packedCheckpoint, err := ci.objdb.RetrievePacked(h.String())
	// TODO: test this case
	if err != nil {
		panic("index: can't retreive packed")
		// return nil, errors.Wrap(err, fmt.Sprintf("can't read object %s", h.String()))
	}

	// TODO: test this case
	if packedCheckpoint.Type() != packing.ObjectTypeCheckpoint {
		panic("index: checkpoint had wrong type after retrieval")
		// return nil, errors.Wrap(err, fmt.Sprintf("object was not a %s but a %s", packing.ObjectTypeCheckpoint, packedCheckpoint.Type()))
	}

	checkpoint, err = jp.UnpackCheckpoint(packedCheckpoint.Contents())
	// TODO: test this case
	if err != nil {
		panic("index: can't unpack checkpoint")
		// return nil, errors.Wrap(err, fmt.Sprintf("can't read object %s", h.String()))
	}

	// If we reached a parentless commit, we are done, lets signal
	// that, and then clean-up
	if len(checkpoint.ParentHashes) == 0 {
		ci.rootReached <- struct{}{}
	}

	for _, ph := range checkpoint.ParentHashes {
		ci.wg.Add(1)
		go ci.retrieveCheckpoint(ph)
	}

	ci.cp <- tocCP{h, timeFromCheckpointFields(checkpoint)}
}

func timeFromCheckpointFields(cp packing.Checkpoint) time.Time {
	t, err := time.Parse(time.RFC3339, cp.Fields["date"])
	if err != nil {
		panic(fmt.Sprint("extremely serious error, could not parse a time we previously validated", err))
	}
	return t
}
