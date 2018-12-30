package depot

import (
	"context"
	"fmt"

	"github.com/golang-collections/collections/stack"
	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/types"
)

// SimpleEventIter emits events on a given partition
type simpleEventIterator struct {
	objdb object.DB

	eventManifest types.EventManifest

	// tipHash is the known "tip" where we started,
	// name is chosen to avoid conflating "head" and "ref"
	// when starting a partition iterator which will start
	// the event iterator, tipHash will be equal to the head
	// ref, but symbolic refs are just for user friendliness
	tipHash types.Hash

	pattern string

	stack   stack.Stack
	matcher PatternMatcher
}

func (s *simpleEventIterator) Pattern() string {
	return s.pattern
}

func (s *simpleEventIterator) TipHash() types.Hash {
	return s.tipHash
}

func (s *simpleEventIterator) Next() types.PersistedEvent {
	return PersistedEv{}
}

func (s *simpleEventIterator) Events(ctx context.Context) (<-chan types.PersistedEvent, <-chan error) {

	var (
		out    = make(chan types.PersistedEvent)
		errOut = make(chan error, 1)
	)

	go func() {

		defer close(out)
		defer close(errOut)

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
					errOut <- fmt.Errorf("error checking partition name %s against pattern %s for match", partitionName, s.pattern)
					return
				}
				if match {
					for _, evHash := range affixEvHashes {

						packedEv, err := s.objdb.RetrievePacked(evHash.String())
						if err != nil {
							// TODO: test me
							errOut <- errors.Wrap(err, "error retrieving packed object from odb from evHash")
							return
						}

						if packedEv.Type() != packing.ObjectTypeEvent {
							// TODO: test me
							errOut <- errors.Wrap(err, fmt.Sprintf("object was not a %s but a %s", packing.ObjectTypeEvent, packedEv.Type()))
							return
						}

						evName, evPayload, err := jp.UnpackEvent(packedEv.Contents())
						if err != nil {
							// TODO: test me
							errOut <- errors.Wrap(err, fmt.Sprintf("can't unpack event %s", packedEv.Contents()))
							return
						}

						pEv := PersistedEv{
							time:          h.time,
							bytes:         evPayload,
							name:          evName,
							partitionName: types.PartitionName(s.pattern),
							cpHash:        h.checkpointHash,
							eventManifest: s.eventManifest,
						}

						select {
						case out <- pEv:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}
	}()

	return out, errOut
}
