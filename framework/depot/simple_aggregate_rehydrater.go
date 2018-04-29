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

type simpleAggregateRehydrater struct {
	objdb object.DB
	refdb ref.DB

	eventManifest types.EventManifest

	pattern types.PartitionName
	matcher PatternMatcher
}

// refs/heads/master => sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824

// Given users/dario
// 1. someone else figured its "users/" ==> &User{} == dario
// 2. go to depo
// 3. get all events for "users/dario"
// 4. loop over them and apply them to &dario

func (s simpleAggregateRehydrater) Rehydrate(ctx context.Context, dst types.Aggregate, partitionName types.PartitionName) error {

	var (
		oStack cpAffixStack
		jp     *packing.JSONPacker
	)

	// Resolve the head ref for the given ctx
	headRef, err := s.refdb.Retrieve(refFromCtx(ctx))
	if err != nil {
		return nil
		return errors.Wrapf(err, "unknown ref, can't lookup partitions for %s", string(partitionName))
	}

	// enqueueCheckpointIfRelevant will push the checkpoint and any ancestors
	// onto the stack and we'll continue when the recursive enqueueCheckpointIfRelevant
	// breaks the loop and we come back here.
	err = s.enqueueCheckpointIfRelevant(headRef, &oStack)
	if err != nil {
		return errors.Wrap(err, "error when stacking relevant partitions")
	}

	for {
		rC := oStack.s.Pop()
		if rC == nil {
			break
		}
		h := rC.(relevantCheckpoint)
		for partitionName, affixEvHashes := range h.affix {
			match, err := s.matcher.DoesMatch(string(partitionName), string(s.pattern))
			if err != nil {
				return fmt.Errorf("error checking partition name %s against pattern %s for match", partitionName, s.pattern)
			}
			if match {
				for _, evHash := range affixEvHashes {

					packedEv, err := s.objdb.RetrievePacked(evHash.String())
					if err != nil {
						// TODO: test me
						fmt.Println("return 1")
						return errors.Wrap(err, "error retrieving packed object from odb from evHash")
					}

					if packedEv.Type() != packing.ObjectTypeEvent {
						// TODO: test me
						fmt.Println("return 2")
						fmt.Printf("object was not a %s but a %s\n", packing.ObjectTypeEvent, packedEv.Type())
						return errors.Wrap(err, fmt.Sprintf("object was not a %s but a %s", packing.ObjectTypeEvent, packedEv.Type()))
					}

					evName, evPayload, err := jp.UnpackEvent(packedEv.Contents())
					if err != nil {
						// TODO: test me
						fmt.Println("return 3")
						return errors.Wrap(err, fmt.Sprintf("can't unpack event %s", packedEv.Contents()))
					}

					_ = evName
					_ = evPayload
					var ev types.Event
					err = dst.ReactTo(ev)

					// fmt.Println(evName, evPayload, h.checkpointHash.String())

					// fmt.Printf("%q before unmarshal %#v\n", evPayload, ev)
					// json.Unmarshal(evPayload, &ev)
					// fmt.Printf("%q after unmarshal %#v\n", evPayload, ev)
				}
			}
		}
	}

	return nil
}

// / enqueueCheckpointIfRelevant pushes checkpoint hashes and affix metadata onto a stack
// which the caller can then drain. enqueueCheckpointIfRelevant is expected to be called
// with a HEAD ref so that the most recent checkpoint on any given thread is pushed onto
// the stack first, and emitted last.
func (s simpleAggregateRehydrater) enqueueCheckpointIfRelevant(checkpointObjHash types.Hash, st *cpAffixStack) error {

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
		matched, err := s.matcher.DoesMatch(string(s.pattern), string(partition))
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
