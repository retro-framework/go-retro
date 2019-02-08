package depot

import (
	"time"

	"github.com/retro-framework/go-retro/framework/types"
)

type PersistedEv struct {
	time  time.Time
	name  string
	bytes []byte

	partitionName types.PartitionName

	// The hash of the checkpoint which referred to the affix
	// from which this event was retrieved/unpacked.
	cpHash types.Hash
}

func (pEv PersistedEv) Time() time.Time {
	return pEv.time
}

func (pEv PersistedEv) Name() string {
	return pEv.name
}

func (pEv PersistedEv) PartitionName() types.PartitionName {
	return pEv.partitionName
}

func (pEv PersistedEv) Bytes() []byte {
	return pEv.bytes
}

func (pEv PersistedEv) CheckpointHash() types.Hash {
	return pEv.cpHash
}
