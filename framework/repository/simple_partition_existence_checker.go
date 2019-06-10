package repository

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/ref"
	"github.com/retro-framework/go-retro/framework/retro"
	"github.com/retro-framework/go-retro/framework/storage"
)

type simplePartitionExistenceChecker struct {
	objdb   object.DB
	refdb   ref.DB
	pattern retro.PartitionName
	matcher retro.Matcher
}

// refs/heads/master => sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824

// Given users/dario
// 1. someone else figured its "users/" ==> &User{} == dario
// 2. go to depo
// 3. get all events for "users/dario"
// 4. loop over them and apply them to &dario

func (s simplePartitionExistenceChecker) Exists(ctx context.Context, partitionName retro.PartitionName) (bool, error) {
	spnExists, ctx := opentracing.StartSpanFromContext(ctx, "simplePartitionExistenceChecker.Exists")
	spnExists.SetTag("partitionName", string(partitionName))
	defer spnExists.Finish()
	headRef, err := s.refdb.Retrieve(refFromCtx(ctx))
	if err != nil {
		spnExists.SetTag("error", err)
		return false, err
	}
	return s.returnTruOnMatching(ctx, headRef)
}

// / enqueueCheckpointIfRelevant pushes checkpoint hashes and affix metadata onto a stack
// which the caller can then drain. enqueueCheckpointIfRelevant is expected to be called
// with a HEAD ref so that the most recent checkpoint on any given thread is pushed onto
// the stack first, and emitted last.
func (s simplePartitionExistenceChecker) returnTruOnMatching(ctx context.Context, checkpointObjHash retro.Hash) (bool, error) {

	var jp *packing.JSONPacker

	// Unpack a Checkpoint
	packedCheckpoint, err := s.objdb.RetrievePacked(checkpointObjHash.String())
	if err != nil {
		// database is likely
		if err == storage.ErrUnknownRef {
			return false, nil
		}
		// TODO: test this case
		return false, errors.Wrap(err, fmt.Sprintf("can't read object %s", checkpointObjHash.String()))
	}

	if packedCheckpoint.Type() != packing.ObjectTypeCheckpoint {
		// TODO: test this case
		return false, errors.Wrap(err, fmt.Sprintf("object was not a %s but a %s", packing.ObjectTypeCheckpoint, packedCheckpoint.Type()))
	}

	checkpoint, err := jp.UnpackCheckpoint(packedCheckpoint.Contents())
	if err != nil {
		// TODO: test this case
		return false, errors.Wrap(err, fmt.Sprintf("can't read object %s", checkpointObjHash.String()))
	}

	// Unpack the Affix
	packedAffix, err := s.objdb.RetrievePacked(checkpoint.AffixHash.String())
	if err != nil {
		// TODO: test this case
		return false, errors.Wrap(err, fmt.Sprintf("retrieve affix %s for checkpoint %s", checkpoint.AffixHash.String(), packedCheckpoint.Hash().String()))
	}

	if packedAffix.Type() != packing.ObjectTypeAffix {
		// TODO: test this case
		return false, errors.Wrap(err, fmt.Sprintf("object was not a %s but a %s", packing.ObjectTypeAffix, packedCheckpoint.Type()))
	}

	affix, err := jp.UnpackAffix(packedAffix.Contents())
	if err != nil {
		// TODO: test this case
		return false, errors.Wrap(err, fmt.Sprintf("unpack affix %s for checkpoint %s", checkpoint.AffixHash.String(), packedCheckpoint.Hash().String()))
	}

	for partition := range affix {
		matched, err := s.matcher.DoesMatch(partition.String())
		if err != nil {
			// TODO: test this case
			return false, errors.Wrap(err, fmt.Sprintf("error checking partition name %s against pattern %s for match", partition, s.pattern))
		}
		if matched.Success() {
			return true, nil
		}
	}

	// TODO: we are not deterministic when parents are ordered, we push stuff in a maybe random order
	// subject possibly to the order the hashes are written by the packer, which I believe to be alphabetic
	// Either way we should peek into a structure and find out which checkpoint is younger and start there
	for _, parentCheckpointHash := range checkpoint.ParentHashes {
		return s.returnTruOnMatching(ctx, parentCheckpointHash)
	}

	return false, nil

}
