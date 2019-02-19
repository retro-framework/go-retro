package depot

import (
	"time"

	"github.com/retro-framework/go-retro/framework/retro"
)

type PersistedEv struct {
	time  time.Time
	name  string
	bytes []byte

	partitionName retro.PartitionName

	// The hash of the checkpoint which referred to the affix
	// from which this event was retrieved/unpacked.
	cpHash retro.Hash
}

func (pEv PersistedEv) Time() time.Time {
	return pEv.time
}

func (pEv PersistedEv) Name() string {
	return pEv.name
}

func (pEv PersistedEv) PartitionName() retro.PartitionName {
	return pEv.partitionName
}

func (pEv PersistedEv) Bytes() []byte {
	return pEv.bytes
}

func (pEv PersistedEv) CheckpointHash() retro.Hash {
	return pEv.cpHash
}
