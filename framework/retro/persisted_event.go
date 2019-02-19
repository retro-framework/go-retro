package retro

import "time"

// PersistedEvent mirrors an event embellishing it with the data
// inferred from its place within the Merkle Tree. The raw data
// returned by Bytes() will need to be handled with an EventManifest
// and a name lookup table to instantiate it back to a real concrete type
type PersistedEvent interface {
	Time() time.Time
	Name() string
	Bytes() []byte
	CheckpointHash() Hash
	PartitionName() PartitionName
}
