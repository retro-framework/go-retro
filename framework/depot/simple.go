package depot

import (
	"context"
	"fmt"

	"github.com/retro-framework/go-retro/framework/packing"

	"github.com/golang-collections/collections/stack"

	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/ref"
	"github.com/retro-framework/go-retro/framework/types"
)

type Simple struct {
	objdb object.DB
	refdb ref.DB
}

func refFromCtx(ctx context.Context) string {
	return "refs/heads/main"
}

// TODO: make this respect the actual value that might come in a context
func (s *Simple) refFromCtx(ctx context.Context) string {
	return "refs/heads/main"
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

// Glob makes the world go round
func (s *Simple) Glob(partition string) types.PartitionIterator {
	return &SimplePartitionIterator{
		objdb:   s.objdb,
		refdb:   s.refdb,
		pattern: partition,
	}
}

type Hash interface {
	String() string
}

type relevantCheckpoint struct {
	checkpointHash packing.Hash
	affix          packing.Affix
}

func (rc relevantCheckpoint) String() string {
	return fmt.Sprintf("Checkpoint: %s (%v)", rc.checkpointHash.String(), rc.affix)
}

type cpAffixStack struct {
	s stack.Stack
}

func (os *cpAffixStack) Push(h relevantCheckpoint) {
	os.s.Push(h)
}

func (os *cpAffixStack) Pop() *relevantCheckpoint {
	v := os.s.Pop()
	if v == nil {
		return nil
	}
	rcp := v.(relevantCheckpoint)
	return &rcp
}

type PatternMatcher interface {
	DoesMatch(pattern, partition string) (bool, error)
}
