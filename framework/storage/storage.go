package storage

import (
	"fmt"
	"time"

	"github.com/golang-collections/collections/stack"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/retro"
)

type AffixStack struct {
	s               stack.Stack
	KnownPartitions []retro.PartitionName
}

type RelevantCheckpoint struct {
	Time           time.Time
	CheckpointHash retro.Hash
	Affix          packing.Affix
}

func (rc RelevantCheckpoint) String() string {
	return fmt.Sprintf("Relevant Checkpoint: %s", rc.CheckpointHash.String())
}

func (os *AffixStack) Len() int {
	return os.s.Len()
}

// Push pushes a relavantCheckpoint onto a stack as we walk
// the object graph. It also maintains a youngest-to-oldest
func (os *AffixStack) Push(rc RelevantCheckpoint) {
	os.s.Push(rc)
	for partitionName := range rc.Affix {
		var partitionNameKnown bool
		for _, kp := range os.KnownPartitions {
			if partitionName == kp {
				partitionNameKnown = true
			}
		}
		if !partitionNameKnown {
			os.KnownPartitions = append(os.KnownPartitions, partitionName)
		}
	}
}

func (os *AffixStack) Pop() *RelevantCheckpoint {
	v := os.s.Pop()
	if v == nil {
		return nil
	}
	rcp := v.(RelevantCheckpoint)
	return &rcp
}
