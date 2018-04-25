package depot

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-collections/collections/stack"
	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/types"
)

// SimpleEventIter emits events on a given partition
type simpleEventIterator struct {
	objdb object.DB

	// tipHash is the known "tip" where we started,
	// name is chosen to avoid conflating "head" and "ref"
	// when starting a partition iterator which will start
	// the event iterator, tipHash will be equal to the head
	// ref, but symbolic refs are just for user friendliness
	tipHash *packing.Hash

	pattern string
	c       chan types.PersistedEvent
	err     chan error

	stack   stack.Stack
	matcher PatternMatcher

	errors []error
}

func (s simpleEventIterator) Pattern() string {
	return s.pattern
}

func (s simpleEventIterator) TipHash() types.Hash {
	return *s.tipHash
}

func (s simpleEventIterator) HasErrors() bool {
	return len(s.errors) > 0
}

func (s simpleEventIterator) Errors() []error {
	return s.errors
}

func (s simpleEventIterator) pushErr(err error) {
	//
	// FERR:
	// 	for {
	// 		select {
	// 		case s.err <- err:
	// 			fmt.Fprintf(os.Stdout, "pushed error %s", err)
	// 			break FERR
	// 		default:
	// 			fmt.Fprintf(os.Stdout, "🍒")
	// 		}
	// 	}
	s.errors = append(s.errors, err)
}

func (s simpleEventIterator) Next() types.PersistedEvent {
	return PersistedEv{}
}

func (s simpleEventIterator) Events(ctx context.Context) (<-chan types.PersistedEvent, error) {

	if s.c == nil {
		s.c = make(chan types.PersistedEvent)
	}

	go func() {
		<-ctx.Done()
		if s.c != nil {
			close(s.c)
			s.c = nil
		}
	}()

	go func() {

		var jp *packing.JSONPacker

		for {
			rC := s.stack.Pop()
			if rC == nil {
				break
			}

			h := rC.(relevantCheckpoint)

			for partitionName, affixEvHashes := range h.affix {
				match, err := s.matcher.DoesMatch(string(partitionName), s.pattern)
				if err != nil {
					s.pushErr(fmt.Errorf("error checking partition name %s against pattern %s for match", partitionName, s.pattern))
				}
				if match {
					for _, evHash := range affixEvHashes {

						time.Sleep(100 * time.Millisecond)
						packedEv, err := s.objdb.RetrievePacked(evHash.String())
						if err != nil {
							// TODO: test me
							s.pushErr(errors.Wrap(err, "error retrieving packed object from odb from evHash"))
						}

						if packedEv.Type() == packing.ObjectTypeEvent {
							// TODO: test me
							s.pushErr(errors.Wrap(err, fmt.Sprintf("object was not a %s but a %s", packing.ObjectTypeEvent, packedEv.Type())))
						}

						evName, evPayload, err := jp.UnpackEvent(packedEv.Contents())
						if err != nil {
							// TODO: test me
							s.pushErr(errors.Wrap(err, fmt.Sprintf("can't unpack event %s", packedEv.Contents())))
						}
						pEv := PersistedEv{
							time:   h.time,
							bytes:  evPayload,
							name:   evName,
							cpHash: h.checkpointHash,
						}
						s.c <- pEv
						// TODO: reintroduce this, it'll show if the consumer is blocking
						// select {
						// case s.c <- pEv:
						// }
					}
				}
			}

			// Signal that we're finished
			if s.c != nil {
				close(s.c)
				s.c = nil
			}

		}
	}()

	return s.c, nil
}
