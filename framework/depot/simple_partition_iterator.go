package depot

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/ref"
	"github.com/retro-framework/go-retro/framework/types"
)

type SimplePartitionIterator struct {
	objdb  object.DB
	refdb  ref.DB
	branch string

	pattern string

	c chan types.EventIterator

	errors []error
}

func (s *SimplePartitionIterator) HasErrors() bool {
	return len(s.errors) > 0
}

func (s *SimplePartitionIterator) Errors() []error {
	return s.errors
}

func (s *SimplePartitionIterator) Pattern() string {
	return s.pattern
}

func (s *SimplePartitionIterator) Next() {
	fmt.Println("next called on simplepartition")
}

// Partitions returns a channel which emits partition event iterators
// which in turn emit events.
//
// TODO: decide how this should handle errors.
// Reminder: errors are values
func (s *SimplePartitionIterator) Partitions(ctx context.Context) (<-chan types.EventIterator, types.CancelFunc) {

	var (
		oStack   cpAffixStack
		cancelFn = func() { close(s.c) }
	)
	if s.c == nil {
		s.c = make(chan types.EventIterator)
	}

	go func() {

		// TODO: Maybe error handling should set an Error on the EventIterator
		// that is returned, that'd be the easiest way to use it, consume all
		// events until it closes, and then check it for errors?
		//
		// EDIT: Scheme in place, seems to be fair enough?

		// Resolve the head ref for the given ctx
		checkpointHash, err := s.refdb.Retrieve(refFromCtx(ctx))
		if err != nil {
			fmt.Println("got here")
			s.errors = append(s.errors, errors.Wrap(err, "unknown reference, can't lookup partitions"))
		}

		err = s.enqueueCheckpointIfRelevant(*checkpointHash, &oStack)
		if err != nil {
			fmt.Println("got here too")
			s.errors = append(s.errors, errors.Wrap(err, "error when stacking relevant partitions"))
		}

		for {
			h := oStack.Pop()
			fmt.Println(h)
			if h == nil {
				break
			}
		}

		// TODO: Now I need to distinct the items on the stack and emit
		// unique event iterators for each partitions I have seen. The
		// structure in oStack might need some work? cpAffixStack may be
		// extendable to maintain a list of seen partitions since Push()
		// has access to the affix metadata on Push()

	}()

	// TODO make something go looking for partitions on the stream
	return s.c, cancelFn
}

// enqueueCheckpointIfRelevant pushes checkpoint hashes and affix metadata onto a stack
// which the caller can then drain. enqueueCheckpointIfRelevant is expected to be called
// with a HEAD ref so that the most recent checkpoint on any given thread is pushed onto
// the stack first, and emitted last.
func (s *SimplePartitionIterator) enqueueCheckpointIfRelevant(checkpointObjHash packing.Hash, st *cpAffixStack) error {

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

	// Check if the Affix matches
	var gpm = GlobPatternMatcher{}
	for partition := range affix {
		matched, err := gpm.DoesMatch(s.pattern, string(partition))
		if err != nil {
			// TODO: test this case
			return errors.Wrap(err, fmt.Sprintf("error checking partition name %s against pattern %s for match", partition, s.pattern))
		}
		if matched {
			st.Push(relevantCheckpoint{
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
