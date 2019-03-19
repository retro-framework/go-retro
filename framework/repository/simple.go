package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/storage/memory"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/framework/matcher"
	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/ref"
	"github.com/retro-framework/go-retro/framework/storage"
	"github.com/retro-framework/go-retro/framework/retro"
)

type simple struct {
	objdb object.DB
	refdb ref.DB

	eventManifest retro.EventManifest

	matcher retro.Matcher
}

type double struct {
	fixture retro.EventFixture
}

func NewEmptyMemory() retro.Repo {
	return simple{
		objdb:         &memory.ObjectStore{},
		refdb:         &memory.RefStore{},
		eventManifest: events.NewManifest(),
	}
}

func NewSimpleRepository(odb object.DB, rdb ref.DB, evM retro.EventManifest) retro.Repo {
	return simple{
		objdb:         odb,
		refdb:         rdb,
		eventManifest: evM,
	}
}

func NewSimpleRepositoryDouble(evFix retro.EventFixture) retro.Repo {
	return double{evFix}
}

func (s double) Claim(ctx context.Context, partition string) bool {
	// TODO: Implement locking properly
	return true
}

func (s double) Release(partition string) {
	// TODO: Implement locking properly
	return
}

func (s double) Exists(_ context.Context, partitionName retro.PartitionName) bool {
	for k := range s.fixture {
		if k.Name() == partitionName {
			return true
		}
	}
	return false
}

func (s double) Rehydrate(ctx context.Context, dst retro.Aggregate, partitionName retro.PartitionName) error {
	var events []retro.Event
	for k, v := range s.fixture {
		if k.Name() == partitionName {
			events = v
		}
	}
	for _, ev := range events {
		dst.ReactTo(ev)
	}
	return nil
}

func (s simple) Claim(ctx context.Context, partition string) bool {
	// TODO: Implement locking properly
	return true
}

func (s simple) Release(partition string) {
	// TODO: Implement locking properly
	return
}

func (s simple) Exists(ctx context.Context, partitionName retro.PartitionName) bool {
	found, _ := simplePartitionExistenceChecker{
		objdb:   s.objdb,
		refdb:   s.refdb,
		pattern: partitionName,
		matcher: matcher.NewGlobPattern(string(partitionName)),
	}.Exists(ctx, partitionName)
	return found
}

func (s simple) Rehydrate(ctx context.Context, dst retro.Aggregate, partitionName retro.PartitionName) error {

	spnRehydrate, ctx := opentracing.StartSpanFromContext(ctx, "repository.simple.Rehydrate")
	spnRehydrate.SetTag("partitionName", string(partitionName))
	defer spnRehydrate.Finish()

	var (
		oStack storage.AffixStack
		jp     *packing.JSONPacker
	)

	// Resolve the head ref for the given ctx
	headRef, err := s.refdb.Retrieve(refFromCtx(ctx))
	if err != nil {
		return errors.Wrapf(err, "unknown ref, can't lookup partitions for %s", string(partitionName))
	}

	spanGatherCheckpoints := opentracing.StartSpan("gathering relevant checkpoints", opentracing.ChildOf(spnRehydrate.Context()))
	defer spanGatherCheckpoints.Finish()
	// enqueueCheckpointIfRelevant will push the checkpoint and any ancestors
	// onto the stack and we'll continue when the recursive enqueueCheckpointIfRelevant
	// breaks the loop and we come back here.
	err = s.enqueueCheckpointIfRelevant(partitionName, headRef, &oStack)
	if err != nil {
		return errors.Wrap(err, "error when stacking relevant partitions")
	}
	spnRehydrate.LogFields(log.Int("found.checkpoints", oStack.Len()))
	spanGatherCheckpoints.Finish()

	spanDrainCheckpoints := opentracing.StartSpan("draining relavant checkpoints", opentracing.ChildOf(spnRehydrate.Context()))
	defer spanDrainCheckpoints.Finish()
	for {
		h := oStack.Pop()
		if h == nil {
			break
		}
		for partitionName, affixEvHashes := range h.Affix {
			match, err := s.matcher.DoesMatch(partitionName)
			if err != nil {
				return fmt.Errorf("error checking partition name %s against pattern %s for match", partitionName, partitionName)
			}
			// the affix contains events for other aggregates
			// but no biggie
			// TODO: remove this guard and test this case(!)
			if partitionName != dst.Name() {
				continue
			}
			if match {

				for _, evHash := range affixEvHashes {

					spanApplyEv := opentracing.StartSpan(
						fmt.Sprintf("apply event %s", evHash.String()),
						opentracing.ChildOf(spanDrainCheckpoints.Context()),
					)
					defer spanApplyEv.Finish()

					spanApplyEv.LogFields(
						log.String("event.hash", evHash.String()),
					)

					packedEv, err := s.objdb.RetrievePacked(evHash.String())
					if err != nil {
						// TODO: test me
						return errors.Wrap(err, "error retrieving packed object from odb from evHash")
					}

					if packedEv.Type() != packing.ObjectTypeEvent {
						// TODO: test me
						return errors.Wrap(err, fmt.Sprintf("object was not a %s but a %s", packing.ObjectTypeEvent, packedEv.Type()))
					}

					evName, evPayload, err := jp.UnpackEvent(packedEv.Contents())
					if err != nil {
						// TODO: test me
						return errors.Wrap(err, fmt.Sprintf("can't unpack event %s", packedEv.Contents()))
					}

					spanApplyEv.LogFields(
						log.String("event.name", evName),
						log.String("event.payload", string(evPayload)),
					)

					ev, err := s.eventManifest.ForName(evName)
					if err != nil {
						return errors.Wrap(err, fmt.Sprintf("can't get event with name %s from manifest", evName))
					}

					err = json.Unmarshal(evPayload, &ev)
					if err != nil {
						return errors.Wrap(err, fmt.Sprintf("can't get unmarshal %s into event registered with name %s: %s", evPayload, evName, err))
					}

					if err = dst.ReactTo(ev); err != nil {
						return errors.Wrap(err, fmt.Sprintf("error applying %q", evName))
					}

					spanApplyEv.Finish()
				}
			}
		}
	}
	spanDrainCheckpoints.Finish()

	return nil
}

// enqueueCheckpointIfRelevant pushes checkpoint hashes and affix metadata onto a stack
// which the caller can then drain. enqueueCheckpointIfRelevant is expected to be called
// with a HEAD ref so that the most recent checkpoint on any given thread is pushed onto
// the stack first, and emitted last.
func (s simple) enqueueCheckpointIfRelevant(pattern retro.PartitionName, checkpointObjHash retro.Hash, st *storage.AffixStack) error {

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
		matched, err := matcher.NewGlobPattern(string(pattern)).DoesMatch(string(partition))
		if err != nil {
			// TODO: test this case
			return errors.Wrap(err, fmt.Sprintf("error checking partition name %s against pattern %s for match", partition, pattern))
		}
		if matched {
			st.Push(storage.RelevantCheckpoint{
				Time:           time.Time{},
				CheckpointHash: packedCheckpoint.Hash(),
				Affix:          affix,
			})
		}
	}

	// TODO: we are not deterministic when parents are ordered, we push stuff in a maybe random order
	// subject possibly to the order the hashes are written by the packer, which I believe to be alphabetic
	// Either way we should peek into a structure and find out which checkpoint is younger and start there
	for _, parentCheckpointHash := range checkpoint.ParentHashes {
		err := s.enqueueCheckpointIfRelevant(pattern, parentCheckpointHash, st)
		if err != nil {
			errors.Wrap(err, fmt.Sprintf("error looking up parent hash %s for checkpoint %s", parentCheckpointHash.String(), packedCheckpoint.Hash().String()))
		}
	}

	return nil
}
