package depot

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/types"
)

// SimpleEventIter emits events on a given partition
type simpleEventIterator struct {
	objdb object.DB

	eventManifest types.EventManifest

	pattern string

	stackCh chan cpAffixStack

	matcher PatternMatcher
}

func (s *simpleEventIterator) Pattern() string {
	return s.pattern
}

func (s *simpleEventIterator) Next() types.PersistedEvent {
	return PersistedEv{}
}

func (s *simpleEventIterator) Events(ctx context.Context) (<-chan types.PersistedEvent, <-chan error) {

	var (
		out    = make(chan types.PersistedEvent)
		errOut = make(chan error, 1)
		jp     *packing.JSONPacker
	)

	var drainStack = func(ctx context.Context, out chan<- types.PersistedEvent, errOut chan<- error, stack cpAffixStack) {
		for {
			var h = stack.Pop()
			if h == nil {
				break
			}
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
							// TODO: metrics (rates?)
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}
	}

	go func() {
		defer close(out)
		defer close(errOut)
		for {
			select {
			case stack := <-s.stackCh:
				drainStack(ctx, out, errOut, stack)
			case <-ctx.Done():
				// fmt.Println("ei: ", ctx.Err())
				return
			}
		}
	}()

	return out, errOut
}
