package index

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/ref"
	"github.com/retro-framework/go-retro/framework/retro"
)

// ChronologicalIndexForBranch returns a streaming index
// for the given branch name.
func ChronologicalForBranch(s string, odb object.DB, refdb ref.DB) (*checkpointChronologiclIndex, func()) {
	var cpCI = &checkpointChronologiclIndex{
		objdb:      odb,
		refdb:      refdb,
		branchName: s,
		toc:        tocCPs{},
		stop:       make(chan struct{}),
		cp:         make(chan tocCP),
		// do not set `ch` here
	}
	var cancelFn = func() {
		cpCI.stop <- struct{}{}
	}
	return cpCI, cancelFn
}

type tocCP struct {
	h retro.Hash
	t time.Time
}

type tocCPs []tocCP

func (r tocCPs) Insert(t tocCP) tocCPs {
	// early return if we know this checkpoint already
	for _, myT := range r {
		if bytes.Equal(t.h.Bytes(), myT.h.Bytes()) {
			return r
		}
	}
	var index = sort.Search(len(r), func(i int) bool {
		return r[i].t.Equal(t.t) || r[i].t.After(t.t)
	})
	return append(r[:index], append([]tocCP{t}, r[index:]...)...)
}

type checkpointChronologiclIndex struct {
	// storage, object and our working slice
	objdb object.DB
	refdb ref.DB

	toc tocCPs

	// branch name
	branchName string

	// "public" api, next will read from ch to
	// emit hashes, stop allows peopel to stop
	// by using the stopfn returned from the helpful
	// constructor
	ch   chan retro.Hash
	stop chan struct{}

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
}

func (ci *checkpointChronologiclIndex) Next(ctx context.Context) retro.Hash {

	// bootstrap if we are nil, having no ci.ch signals that we
	// have no index yet, and aren't running.
	if ci.ch == nil {
		go ci.run()

		ref, err := ci.refdb.Retrieve(refFromCtx(ctx))
		if err != nil {
			panic("index: error retrieving ref")
		}
		ci.wg.Add(1)
		ci.retrieveCheckpoint(ref)
	}

	ci.wg.Wait()
	ci.ch = make(chan retro.Hash)

	defer func() { ci.cursor++ }()
	return ci.toc[ci.cursor].h
}

func (ci *checkpointChronologiclIndex) run() {
	for {
		select {
		case t := <-ci.cp:
			ci.toc = ci.toc.Insert(t)
			ci.wg.Done()
		case <-ci.stop:
			// TODO handle stop case properly?
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

	for _, ph := range checkpoint.ParentHashes {
		ci.wg.Add(1)
		go ci.retrieveCheckpoint(ph)
	}

	ci.cp <- tocCP{h, timeFromCheckpointFields(checkpoint)}
}

// TODO: make this respect the actual value that might come in a context
func refFromCtx(ctx context.Context) string {
	// TODO:  this code now duplicated in at least two places
	return "refs/heads/master"
}

func timeFromCheckpointFields(cp packing.Checkpoint) time.Time {
	t, err := time.Parse(time.RFC3339, cp.Fields["date"])
	if err != nil {
		fmt.Println("extremely serious error, could not parse a time we previously validated", err)
	}
	return t
}
