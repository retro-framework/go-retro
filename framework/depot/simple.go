package depot

import (
	"context"
	"fmt"

	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/ref"
	"github.com/retro-framework/go-retro/framework/types"
)

type Simple struct {
	odb object.DB
	rdb ref.DB
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

type SimplePartitionIterator struct {
}

func (s *SimplePartitionIterator) Next() {
	fmt.Println("next called on simplepartition")
}

// func (g *Simple) Glob()

// Depot stores events, commands, etc. It is heavily inspired
// by Git's model of generic object and ref stores linked with
// pointers. It's aim is to be correct, not fast. To be verifiable,
// and duplicable.
// type Depot interface {
//
// 	// StoreEvent takes a domain event and returns a Hash
// 	// the must be deterministic and not affected by PRNG
// 	// or types, just the serialization format (repository
// 	// is a storage concern)
// 	StoreEvent(Event) (Hash, error)
//
// 	// StoreAffix stores an affix. An affix may contain
// 	// a new set of events for one or more partitions, given
// 	// that we know the name of the aggregate being changed
// 	// most affixes will contain one partition name and one
// 	// or more event hashes.
// 	StoreAffix(Affix) (Hash, error)
//
// 	// StoreCheckpoint stores a checkpoint. A checkpoint
// 	// is approximately equivilant to a Git commit. An
// 	// object under heavy writes however may "auto branch"
// 	// checkpoints (multiple checkpoints with a common parent)
// 	// which we will have to resolve deterministically later
// 	StoreCheckpoint(Checkpoint) (Hash, error)
// }
