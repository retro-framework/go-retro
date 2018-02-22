package depot

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
