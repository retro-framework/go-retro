package depot

import (
	"context"
	"fmt"
	"time"

	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/ref"
	"github.com/retro-framework/go-retro/framework/types"
)

type Simple struct {
	objdb object.DB
	refdb ref.DB
}

// TODO: make this respect the actual value that might come in a context
func (s *Simple) refFromCtx(ctx context.Context) string {
	return "refs/heads/mainline"
}

func (s *Simple) Claim(ctx context.Context, partition string) bool {
	// TODO: Implement locking properly
	return true
}

func (s *Simple) Release(partition string) {
	// TODO: Implement locking properly
	return
}

// Rehydrate
func (s *Simple) Rehydrate(partition string, dst types.Aggregate, sid types.SessionID) error {
	return nil
}

// Glob
func (s *Simple) Glob(partition string) types.PartitionIterator {
	// TODO: Implement locking properly
	return &SimplePartitionIterator{}
}

func (s *Simple) GetEvents(partition string) types.PartitionIterator {
	return &SimplePartitionIterator{pattern: partition}
}

type SimplePartitionEvent struct{}

func (s *SimplePartitionEvent) Time() time.Time {
	return time.Time{}
}

func (s *SimplePartitionEvent) Name() string {
	return "demo"
}

func (s *SimplePartitionEvent) Bytes() []byte {
	return []byte{}
}

type SimplePartitionIterator struct {
	pattern string
	c       chan types.EventIterator
}

func (s *SimplePartitionIterator) Pattern() string {
	return s.pattern
}

func (s *SimplePartitionIterator) Next() {
	fmt.Println("next called on simplepartition")
}

// Partitions returns a channel which emits partition event iterators
// which in turn emit events
func (s *SimplePartitionIterator) Partitions() (<-chan types.EventIterator, types.CancelFunc) {
	if s.c == nil {
		s.c = make(chan types.EventIterator)
	}
	// TODO make something go looking for partitions on the stream
	return s.c, func() { close(s.c) }
}

// SimpleEventIter emits events on a given partition
type SimpleEventIterator struct {
	pattern string
	c       chan types.PersistedEvent
}

func (s *SimpleEventIterator) Pattern() string {
	return s.pattern
}

func (s *SimpleEventIterator) Next() {
	fmt.Println("next called on simplepartition")
}

func (s *SimpleEventIterator) Events() (<-chan types.PersistedEvent, types.CancelFunc) {
	if s.c == nil {
		s.c = make(chan types.PersistedEvent)
	}
	return s.c, func() { close(s.c) }
}
