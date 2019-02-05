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

	subscribedOn <-chan types.RefMove

	eventIterators map[string]types.EventIterator
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
		out    = make(chan types.EventIterator)
		stacks = make(chan cpAffixStack)
		errOut = make(chan error, 1)
	)

	var emitPartitionIterator = func(ctx context.Context, out chan<- types.EventIterator, oStack cpAffixStack, kp string) {
		// Check if we have a consumer for this
		// already, doesn't check if that consumer
		// is still behaving properly
		// if _, ok := s.eventIterators[kp]; ok {
		// 	return
		// }
		evIter := &simpleEventIterator{
			objdb:         s.objdb,
			matcher:       s.matcher,
			eventManifest: s.eventManifest,
			pattern:       kp,
			tipHash:       s.tipHash,
			stack:         oStack.s, // a *copy* of the stack, so we don't mutate it
		}
		select {
		case out <- evIter:
			s.eventIterators[kp] = evIter
		case <-ctx.Done():
			return
		}
	}

	var collectRelevantCheckpoints = func(from, to types.Hash) error {
		var st cpAffixStack
		// enqueueCheckpointIfRelevant will push the checkpoint and any ancestors
		// onto the stack and we'll continue when the recursive enqueueCheckpointIfRelevant
		// breaks the loop and we come back here.
		//
		// TODO: make this respect "from" by never traversing too far backwards
		var err = s.enqueueCheckpointIfRelevant(from, to, &st)
		if err != nil {
			return errors.Wrap(err, "error when stacking relevant partitions")
		}
		stacks <- st
		return nil
	}

	go func() {

		defer func() {
			close(out)
			close(errOut)
		}()

		// Resolve the head ref for the given ctx
		checkpointHash, err := s.refdb.Retrieve(refFromCtx(ctx))
		if err != nil {
			errOut <- errors.Wrap(err, "unknown reference, can't lookup partitions")
			return
		}

		go func() {
			err = collectRelevantCheckpoints(nil, checkpointHash)
			if err != nil {
				errOut <- errors.Wrap(err, "unknown reference, can't lookup partitions")
				return
			}
		}()

		for {
			select {
			case newStack, ok := <-stacks:
				if ok {
					for _, kp := range newStack.knownPartitions {
						emitPartitionIterator(ctx, out, newStack, string(kp))
					}
				}
			case refMoved, ok := <-s.subscribedOn:
				if ok {
					go collectRelevantCheckpoints(refMoved.Old, refMoved.New)
				}
			case <-ctx.Done():
				return
			}
		}

	}()

	return out, errOut
}

// enqueueCheckpointIfRelevant pushes checkpoint hashes and affix metadata onto a stack
// which the caller can then drain. enqueueCheckpointIfRelevant is expected to be called
// with a HEAD ref so that the most recent checkpoint on any given thread is pushed onto
// the stack first, and emitted last.
func (s *simplePartitionIterator) enqueueCheckpointIfRelevant(fromHash, toHash types.Hash, st *cpAffixStack) error {

	var jp *packing.JSONPacker

	// Unpack a Checkpoint
	packedCheckpoint, err := s.objdb.RetrievePacked(toHash.String())
	if err != nil {
		// TODO: test this case
		return errors.Wrap(err, fmt.Sprintf("can't read object %s", toHash.String()))
	}

	if packedCheckpoint.Type() != packing.ObjectTypeCheckpoint {
		// TODO: test this case
		return errors.Wrap(err, fmt.Sprintf("object was not a %s but a %s", packing.ObjectTypeCheckpoint, packedCheckpoint.Type()))
	}

	checkpoint, err := jp.UnpackCheckpoint(packedCheckpoint.Contents())
	if err != nil {
		// TODO: test this case
		return errors.Wrap(err, fmt.Sprintf("can't read object %s", toHash.String()))
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

			// Ensure we can get the date field header and parse it, else raise an error.
			dateStr, ok := checkpoint.Fields["date"]
			if !ok {
				// TODO: test this case
				return fmt.Errorf("error retrieving date field from checkpoint fields (checkpoint hash %s)", toHash.String())
			}

			t, err := time.Parse(time.RFC3339, dateStr)
			if err != nil {
				// TODO: test this case
				return errors.Wrap(err, fmt.Sprintf("parsing date %q as rfc3339", dateStr))
			}

			st.Push(relevantCheckpoint{
				time:           t,
				checkpointHash: packedCheckpoint.Hash(),
				affix:          affix,
			})
		}
	}

	// TODO: we are not deterministic when parents are ordered, we push stuff in a maybe random order
	// subject possibly to the order the hashes are written by the packer, which I believe to be alphabetic
	// Either way we should peek into a structure and find out which checkpoint is younger and start there
	for _, parentCheckpointHash := range checkpoint.ParentHashes {
		// early return, we've come as far back in the ancestry
		// as we were asked.
		if parentCheckpointHash != nil && fromHash != nil {
			if parentCheckpointHash.String() == fromHash.String() {
				return nil
			}
		}
		err := s.enqueueCheckpointIfRelevant(fromHash, parentCheckpointHash, st)
		if err != nil {
			errors.Wrap(err, fmt.Sprintf("error looking up parent hash %s for checkpoint %s", parentCheckpointHash.String(), packedCheckpoint.Hash().String()))
		}
	}

	return nil
}
