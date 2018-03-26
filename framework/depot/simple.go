package depot

import (
	"context"

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
	checkpointHash Hash
}

type cpAffixStack struct{ h []Hash }

func (os *cpAffixStack) Push(h Hash) { os.h = append(os.h, h) }
func (os *cpAffixStack) Pop() Hash {
	var x Hash
	x, os.h = os.h[0], os.h[1:]
	return x
}

type PatternMatcher interface {
	DoesMatch(pattern, partition string) (bool, error)
}
