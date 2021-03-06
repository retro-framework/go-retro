package depot

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/retro"
	"github.com/retro-framework/go-retro/framework/storage"
)

// SimpleEventIter emits events on a given partition
type simpleEventIterator struct {
	objdb object.DB

	pattern string

	stackCh chan storage.AffixStack

	matcher retro.Matcher

	out    chan retro.PersistedEvent
	outErr chan error
}

func (s *simpleEventIterator) Pattern() string {
	return s.pattern
}

func (s *simpleEventIterator) Next(ctx context.Context) (retro.PersistedEvent, error) {

	if s.out == nil && s.outErr == nil {
		s.out = make(chan retro.PersistedEvent)
		s.outErr = make(chan error, 1)
		go s.events(ctx, s.out, s.outErr)
	}

	select {
	case evIter := <-s.out:
		return evIter, nil
	case err := <-s.outErr:
		return nil, err
	case <-ctx.Done():
		return nil, Done
	}
}

func (s *simpleEventIterator) Events(ctx context.Context) (<-chan retro.PersistedEvent, <-chan error) {
	var (
		out    = make(chan retro.PersistedEvent)
		outErr = make(chan error, 1)
	)
	return s.events(ctx, out, outErr)
}

func (s *simpleEventIterator) events(ctx context.Context, out chan retro.PersistedEvent, outErr chan error) (<-chan retro.PersistedEvent, <-chan error) {

	var jp *packing.JSONPacker

	var drainStack = func(ctx context.Context, out chan<- retro.PersistedEvent, outErr chan<- error, stack storage.AffixStack) {
		for {
			var h = stack.Pop()
			if h == nil {
				break
			}
			for partitionName, affixEvHashes := range h.Affix {
				match, err := s.matcher.DoesMatch(partitionName)
				if err != nil {
					outErr <- fmt.Errorf("error checking partition name %s against pattern %s for match", partitionName, s.pattern)
					return
				}
				if match {
					for _, evHash := range affixEvHashes {

						packedEv, err := s.objdb.RetrievePacked(evHash.String())
						if err != nil {
							// TODO: test me
							outErr <- errors.Wrap(err, "error retrieving packed object from odb from evHash")
							return
						}

						if packedEv.Type() != packing.ObjectTypeEvent {
							// TODO: test me
							outErr <- errors.Wrap(err, fmt.Sprintf("object was not a %s but a %s", packing.ObjectTypeEvent, packedEv.Type()))
							return
						}

						evName, evPayload, err := jp.UnpackEvent(packedEv.Contents())
						if err != nil {
							// TODO: test me
							outErr <- errors.Wrap(err, fmt.Sprintf("can't unpack event %s", packedEv.Contents()))
							return
						}

						pEv := PersistedEv{
							time:          h.Time,
							bytes:         evPayload,
							name:          evName,
							partitionName: retro.PartitionName(s.pattern),
							cpHash:        h.CheckpointHash,
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
		defer close(outErr)
		for {
			select {
			case stack := <-s.stackCh:
				drainStack(ctx, out, outErr, stack)
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, outErr
}
