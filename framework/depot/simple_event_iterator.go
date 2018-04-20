package depot

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-collections/collections/stack"
	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/types"
)

type PersistedEv struct {
	time  time.Time
	name  string
	bytes []byte
}

func (pEv PersistedEv) Time() time.Time {
	return pEv.time
}

func (pEv PersistedEv) Name() string {
	return pEv.name
}

func (pEv PersistedEv) Bytes() []byte {
	return pEv.bytes
}

// SimpleEventIter emits events on a given partition
type simpleEventIterator struct {
	objdb object.DB

	pattern string
	c       chan types.PersistedEvent

	stack   stack.Stack
	matcher PatternMatcher

	errors []error
}

func (s simpleEventIterator) Pattern() string {
	return s.pattern
}

func (s simpleEventIterator) HasErrors() bool {
	return len(s.errors) > 0
}

func (s simpleEventIterator) Errors() []error {
	return s.errors
}

func (s simpleEventIterator) Next() types.PersistedEvent {
	return PersistedEv{}
}

func (s simpleEventIterator) Events(ctx context.Context) (<-chan types.PersistedEvent, error) {

	// var (
	// 	cancelFn = func() {
	// 		close(s.c)
	// 		s.c = nil
	// 	}
	// )

	if s.c == nil {
		s.c = make(chan types.PersistedEvent)
	}

	go func() {
		<-ctx.Done()
		close(s.c)
		s.c = nil
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
					s.errors = append(s.errors, fmt.Errorf("error checking partition name %s against pattern %s for match", partitionName, s.pattern))
				}
				if match {
					for _, evHash := range affixEvHashes {
						time.Sleep(100 * time.Millisecond)
						packedEv, err := s.objdb.RetrievePacked(evHash.String())
						if err != nil {
							// TODO: fixme
						}
						if packedEv.Type() == packing.ObjectTypeEvent {
							// TODO: fixme
							// return errors.Wrap(err, fmt.Sprintf("object was not a %s but a %s", packing.ObjectTypeCheckpoint, packedCheckpoint.Type()))
						}
						evName, evPayload, err := jp.UnpackEvent(packedEv.Contents())
						if err != nil {
							// TODO: fixme
							// return errors.Wrap(err, fmt.Sprintf("can't read object %s", checkpointObjHash.String()))
						}
						pEv := PersistedEv{
							time:  h.time,
							bytes: evPayload,
							name:  evName,
						}
						s.c <- pEv
						// TODO: reintroduce this, it'll show if the consumer is blocking
						// select {
						// case s.c <- pEv:
						// }
					}
				}
			}
		}
	}()

	return s.c, nil
}
