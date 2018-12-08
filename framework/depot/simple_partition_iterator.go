package depot

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/ref"
	"github.com/retro-framework/go-retro/framework/types"
)

type simplePartitionIterator struct {
	objdb object.DB
	refdb ref.DB

	eventManifest types.EventManifest

	// tipHash is the known "tip" where we started,
	// name is chosen to avoid conflating "head" and "ref"
	// when starting a partition iterator which will start
	// the event iterator, tipHash will be equal to the head
	// ref, but symbolic refs are just for user friendliness
	tipHash types.Hash

	pattern string
	matcher PatternMatcher
}

func (s *simplePartitionIterator) Pattern() string {
	return s.pattern
}

// Partitions returns a channel which emits partition event iterators
// which in turn emit events.
//
// TODO: decide how this should handle errors.
// Reminder: errors are values
func (s *simplePartitionIterator) Partitions(ctx context.Context) (<-chan types.EventIterator, <-chan error) {

	var (
		oStack cpAffixStack
		out    = make(chan types.EventIterator)
		errOut = make(chan error, 1)
	)

	go func() {

		defer close(out)
		defer close(errOut)

		// Resolve the head ref for the given ctx
		checkpointHash, err := s.refdb.Retrieve(refFromCtx(ctx))
		if err != nil {
			errOut <- errors.Wrap(err, "unknown reference, can't lookup partitions")
			return
		}

		if s.tipHash == nil {
			s.tipHash = checkpointHash
		}

		// enqueueCheckpointIfRelevant will push the checkpoint and any ancestors
		// onto the stack and we'll continue when the recursive enqueueCheckpointIfRelevant
		// breaks the loop and we come back here.
		err = s.enqueueCheckpointIfRelevant(checkpointHash, &oStack)
		if err != nil {
			errOut <- errors.Wrap(err, "error when stacking relevant partitions")
			return
		}

		// TODO: If we get here with no knownPartitions we never continue
		// we should test for that.

		// At this point the oStack should contain all Checkpoints that
		// historically contained an Affix which contained a partition
		// matching the pattern.
		//
		// We can use oStack.knownPartitions to emit the event emitters
		// one, each for each partition. We give the SimpleEventIterator
		// a copy of the oStack that we built, so it can pop them and match
		// things against it's own pattern and sent them to the consumer.
		for _, kp := range oStack.knownPartitions {

			evIter := &simpleEventIterator{
				objdb:         s.objdb,
				matcher:       s.matcher,
				eventManifest: s.eventManifest,
				pattern:       string(kp),
				tipHash:       s.tipHash,
				stack:         oStack.s, // a copy of the stack, so we don't mutate it (?)
			}
			select {
			case out <- evIter:
			case <-ctx.Done():
				return
			}
		}
	}()

	// TODO: make something go looking for partitions on the stream
	// maybe register a callback on the refdb to get changes in HEAD
	// pointer?

	return out, errOut
}

// enqueueCheckpointIfRelevant pushes checkpoint hashes and affix metadata onto a stack
// which the caller can then drain. enqueueCheckpointIfRelevant is expected to be called
// with a HEAD ref so that the most recent checkpoint on any given thread is pushed onto
// the stack first, and emitted last.
func (s *simplePartitionIterator) enqueueCheckpointIfRelevant(checkpointObjHash types.Hash, st *cpAffixStack) error {

	var jp *packing.JSONPacker

	// Unpack a Checkpoint
	packedCheckpoint, err := s.objdb.RetrievePacked(checkpointObjHash.String())
	if err != nil {
		// TODO: test this case
		return errors.Wrap(err, fmt.Sprintf("can't read object %s", checkpointObjHash.String()))
	}

	if packedCheckpoint.Type() != packing.ObjectTypeCheckpoint {
		// TODO: test this case
		return errors.Wrap(err, fmt.Sprintf("object was not a %s but a %s", packing.ObjectTypeCheckpoint, packedCheckpoint.Type()))
	}

	checkpoint, err := jp.UnpackCheckpoint(packedCheckpoint.Contents())
	if err != nil {
		// TODO: test this case
		return errors.Wrap(err, fmt.Sprintf("can't read object %s", checkpointObjHash.String()))
	}

	// Unpack the Affix
	packedAffix, err := s.objdb.RetrievePacked(checkpoint.AffixHash.String())
	if err != nil {
		// TODO: test this case
		return errors.Wrap(err, fmt.Sprintf("retrieve affix %s for checkpoint %s", checkpoint.AffixHash.String(), packedCheckpoint.Hash().String()))
	}

	if packedAffix.Type() != packing.ObjectTypeAffix {
		// TODO: test this case
		return errors.Wrap(err, fmt.Sprintf("object was not a %s but a %s", packing.ObjectTypeAffix, packedCheckpoint.Type()))
	}

	affix, err := jp.UnpackAffix(packedAffix.Contents())
	if err != nil {
		// TODO: test this case
		return errors.Wrap(err, fmt.Sprintf("unpack affix %s for checkpoint %s", checkpoint.AffixHash.String(), packedCheckpoint.Hash().String()))
	}

	for partition := range affix {
		matched, err := s.matcher.DoesMatch(s.pattern, string(partition))
		if err != nil {
			// TODO: test this case
			return errors.Wrap(err, fmt.Sprintf("error checking partition name %s against pattern %s for match", partition, s.pattern))
		}
		if matched {
			st.Push(relevantCheckpoint{
				// TODO: Something about times????
				time:           time.Time{},
				checkpointHash: packedCheckpoint.Hash(),
				affix:          affix,
			})
		}
	}

	// TODO: we are not deterministic when parents are ordered, we push stuff in a maybe random order
	// subject possibly to the order the hashes are written by the packer, which I believe to be alphabetic
	// Either way we should peek into a structure and find out which checkpoint is younger and start there
	for _, parentCheckpointHash := range checkpoint.ParentHashes {
		err := s.enqueueCheckpointIfRelevant(parentCheckpointHash, st)
		if err != nil {
			errors.Wrap(err, fmt.Sprintf("error looking up parent hash %s for checkpoint %s", parentCheckpointHash.String(), packedCheckpoint.Hash().String()))
		}
	}

	return nil
}
