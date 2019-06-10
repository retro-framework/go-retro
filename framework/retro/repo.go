package retro

import "context"

// Repository frames all the behaviours we need to play back events
// and have a domain model which is unconcerned with storage and
// simply behaves as though the entities exist readily in memory.
type Repo interface {
	Claim(context.Context, string) bool
	Release(string)

	Exists(context.Context, PartitionName) bool

	Rehydrate(context.Context, Aggregate, PartitionName) error
}
