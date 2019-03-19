package depot

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/framework/matcher"
	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/ref"
	"github.com/retro-framework/go-retro/framework/retro"
	"github.com/retro-framework/go-retro/framework/storage"
)

var Done = fmt.Errorf("iterator depleted: done")

type simplePartitionIterator struct {
	objdb object.DB
	refdb ref.DB

	pattern string
	matcher retro.Matcher

	subscribedOn <-chan retro.RefMove

	eventIterators map[string]*simpleEventIterator

	out    chan retro.EventIterator
	outErr chan error
}

func (s *simplePartitionIterator) Pattern() string {
	return s.pattern
}

func (s *simplePartitionIterator) Next(ctx context.Context) (retro.EventIterator, error) {

	if s.out == nil && s.outErr == nil {
		s.out = make(chan retro.EventIterator)
		s.outErr = make(chan error, 1)
		go s.partitions(ctx, s.out, s.outErr)
	}

	select {
	case evIter := <-s.out:
		return evIter, nil
	case err := <-s.outErr:
		return nil, err
	case <-ctx.Done():
		return nil, Done
	}
}

func (s *simplePartitionIterator) Partitions(ctx context.Context) (<-chan retro.EventIterator, <-chan error) {
	var (
		out    = make(chan retro.EventIterator)
		outErr = make(chan error, 1)
	)
	return s.partitions(ctx, out, outErr)
}

func (s *simplePartitionIterator) partitions(ctx context.Context, out chan retro.EventIterator, outErr chan error) (<-chan retro.EventIterator, <-chan error) {

	var stacks = make(chan storage.AffixStack)

	var emitPartitionIterator = func(ctx context.Context, out chan<- retro.EventIterator, oStack storage.AffixStack, kp string) {
		// Check if we have a consumer for this
		// already, doesn't check if that consumer
		// is still behaving properly
		if existingEvIter, ok := s.eventIterators[kp]; ok {
			existingEvIter.stackCh <- oStack
			return
		}
		//
		// stackCh is buffered because without the ability to
		// enqueue one stack before the partition consumer
		// begins to process the stack we would block already
		// before being able to return from this function.
		//
		evIter := &simpleEventIterator{
			objdb:   s.objdb,
			matcher: matcher.NewGlobPattern(kp),
			pattern: kp,
			stackCh: make(chan storage.AffixStack, 1),
		}
		select {
		case out <- evIter:
			evIter.stackCh <- oStack
			s.eventIterators[kp] = evIter
		case <-ctx.Done():
			return
		}
	}

	var collectRelevantCheckpoints = func(from, to retro.Hash) error {
		var st storage.AffixStack
		// enqueueCheckpointIfRelevant will push the checkpoint and any ancestors
		// onto the stack and we'll continue when the recursive enqueueCheckpointIfRelevant
		// breaks the loop and we come back here.
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
			close(outErr)
		}()

		// Resolve the head ref for the given ctx
		checkpointHash, err := s.refdb.Retrieve(refFromCtx(ctx))
		if err != nil {
			outErr <- errors.Wrap(err, "unknown reference, can't lookup partitions")
			return
		}

		go func() {
			err = collectRelevantCheckpoints(nil, checkpointHash)
			if err != nil {
				outErr <- errors.Wrap(err, "unknown reference, can't lookup partitions")
				return
			}
		}()

		for {
			select {
			case newStack, ok := <-stacks:
				if ok {
					for _, kp := range newStack.KnownPartitions {
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

	return out, outErr
}

// enqueueCheckpointIfRelevant pushes checkpoint hashes and affix metadata onto a stack
// which the caller can then drain. enqueueCheckpointIfRelevant is expected to be called
// with a HEAD ref so that the most recent checkpoint on any given thread is pushed onto
// the stack first, and emitted last.
func (s *simplePartitionIterator) enqueueCheckpointIfRelevant(fromHash, toHash retro.Hash, st *storage.AffixStack) error {

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
		matched, err := s.matcher.DoesMatch(partition)
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

			st.Push(storage.RelevantCheckpoint{
				Time:           t,
				CheckpointHash: packedCheckpoint.Hash(),
				Affix:          affix,
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
