package depot

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/framework/matcher"
	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/ref"
	"github.com/retro-framework/go-retro/framework/retro"
	"github.com/retro-framework/go-retro/framework/storage"
	"github.com/retro-framework/go-retro/framework/storage/memory"
)

// DefaultBranchName is defined so that without an override changes
// will move the ref named by this branch name
const DefaultBranchName = "refs/heads/master"

// NewSimpleStub returns a simple Depot stub which will yield the given events in the fixture
// as a single checkpoint with a single affix with a generic set of placeholder metadata.
func NewSimpleStub(t *testing.T,
	objDB object.DB,
	refDB ref.DB,
	fixture map[string][]retro.EventNameTuple,
) retro.Depot {
	var (
		jp    = packing.NewJSONPacker()
		affix = packing.Affix{}
	)
	for aggName, evNameTuples := range fixture {
		var (
			evHashesForAffix []retro.Hash
		)
		for _, evNameTuple := range evNameTuples {
			packedEv, err := jp.PackEvent(evNameTuple.Name, evNameTuple.Event)
			if err != nil {
				t.Errorf("error packing event in NewSimpleStub: %s", err)
			}
			if _, err := objDB.WritePacked(packedEv); err != nil {
				t.Errorf("error writing packedEv to odb in NewSimpleStub")
			}
			evHashesForAffix = append(evHashesForAffix, packedEv.Hash())
		}
		affix[retro.PartitionName(aggName)] = evHashesForAffix
	}
	packedAffix, err := jp.PackAffix(affix)
	if err != nil {
		t.Errorf("error packing affix in NewSimpleStub: %s", err)
	}
	if _, err := objDB.WritePacked(packedAffix); err != nil {
		t.Errorf("error writing packedAffix to odb in NewSimpleStub")
	}
	checkpoint := packing.Checkpoint{
		AffixHash:   packedAffix.Hash(),
		CommandDesc: []byte(`{"stub":"article"}`),
		Fields:      map[string]string{"session": "hello world"},
	}
	packedCheckpoint, err := jp.PackCheckpoint(checkpoint)
	if err != nil {
		t.Errorf("error packing checkpoint in NewSimpleStub: %s", err)
	}
	if _, err := objDB.WritePacked(packedCheckpoint); err != nil {
		t.Errorf("error writing packedCheckpoint to odb in NewSimpleStub")
	}
	refDB.Write(DefaultBranchName, packedCheckpoint.Hash())
	return &Simple{objdb: objDB, refdb: refDB}
}

func NewSimple(odb object.DB, refdb ref.DB) retro.Depot {
	return &Simple{objdb: odb, refdb: refdb}
}

// EmptySimpleMemory returns an empty depot to keep the type system happy
func EmptySimpleMemory() retro.Depot {
	return &Simple{
		objdb: &memory.ObjectStore{},
		refdb: &memory.RefStore{},
	}
}

// HeadPointer is a simple read-thru which gets
// the value of the current jhead pointer from the refdb
// it uses the context to try and get a branch name, and
// in case of failure falls back to the default branch
// name.
func (s *Simple) HeadPointer(ctx context.Context) (retro.Hash, error) {
	ptr, err := s.refdb.Retrieve(refFromCtx(ctx))
	if err == storage.ErrUnknownRef {
		return nil, nil
	}
	return ptr, err
}

// DumpAll lists all refs and objects in relative cleartext.
// useful for tests and not much else.
func (s *Simple) DumpAll(w io.Writer) string {
	if lodb, ok := s.objdb.(object.ListableSource); ok {
		var hashStrings []string
		for _, hash := range lodb.Ls() {
			hashStrings = append(hashStrings, hash.String())
		}
		sort.Strings(hashStrings)
		for _, hashStr := range hashStrings {
			ho, err := s.objdb.RetrievePacked(hashStr)
			if err != nil {
				fmt.Fprintln(w, "error retrieving", hashStr)
			}
			fmt.Fprintf(w, "%s:%s\n", ho.Type(), hashStr)
			fmt.Fprintf(w, "%s\n\n", strings.Replace(string(ho.Contents()), "\u0000", "\\u0000", -1))
		}
	}
	if lrefdb, ok := s.refdb.(ref.ListableStore); ok {
		m, err := lrefdb.Ls()
		if err != nil {
			fmt.Fprintf(w, "err: %s", err)
		}
		for k, v := range m {
			fmt.Fprintf(w, "%s -> %s\n", k, v)
		}
	}
	return ""
}

// Simple is the simplest possible Depot implementation
// it requires only a object and ref database implementation
// and an event manifest to map the events from the object
// db to a the time they are restored.
type Simple struct {
	objdb object.DB
	refdb ref.DB

	subscribers []chan<- retro.RefMove
}

// TODO: make this respect the actual value that might come in a context
func refFromCtx(ctx context.Context) string {
	return DefaultBranchName
}

// Watch makes the world go round
func (s *Simple) Watch(_ context.Context, partition string) retro.PartitionIterator {
	var subscriberNotificationCh = make(chan retro.RefMove)
	s.subscribers = append(s.subscribers, subscriberNotificationCh)
	return &simplePartitionIterator{
		objdb:          s.objdb,
		refdb:          s.refdb,
		pattern:        partition,
		matcher:        matcher.NewGlobPattern(partition),
		subscribedOn:   subscriberNotificationCh,
		eventIterators: make(map[string]*simpleEventIterator),
	}
}

// StorePacked takes a variable number of hashed objects, packs and stores them
// in the object store backing the Simple Depot
func (s Simple) StorePacked(packed ...retro.HashedObject) error {
	for _, p := range packed {
		_, err := s.objdb.WritePacked(p)
		if err != nil {
			return errors.Wrap(err, "can't store packed")
		}
	}
	return nil
}

// MoveHeadPointer overwrites the DefaultBranchName with
// the new reference given. It needs to be made safer, see TODO.
//
// TODO: check for fastforward 🔜 before allowing write and/or something
// to make this not totally unsafe
func (s Simple) MoveHeadPointer(old, new retro.Hash) error {
	_, err := s.refdb.Write(DefaultBranchName, new)
	if err == nil {
		s.notifySubscribers(old, new)
	}
	return err
}

func (s Simple) Matching(ctx context.Context, m retro.Matcher) (retro.MatcherResults, error) {
	return queryable{
		objdb: s.objdb,
		refdb: s.refdb,
	}.Matching(ctx, m)
}

// notifySubscribers takes old,new so that we can notify subscribers whether
// this is fast forward or not. That _should_ be as easy as fetching the
// new from the store, and checking that it has the old one as its only
// parent hash.
//
// Probably significant that this receives a _copy_ of subscribers, as the
// receiver is not a pointer type.
//
// It also comes to my mind whether subscribers can be global, or whether
// they need to be differentiated by which pattern they searched for
// I suspect "global" (to the Depot instance) is ok for the time being.
func (s Simple) notifySubscribers(old, new retro.Hash) error {
	for _, subscriber := range s.subscribers {
		go func(subscriber chan<- retro.RefMove) {
			select {
			case subscriber <- retro.RefMove{Old: old, New: new}:
				// TODO: something about metrics ?
			case <-time.After(1 * time.Minute):
				fmt.Fprintf(os.Stderr, "blocked for one minute waiting to notify subscriber, skipping.")
			}
		}(subscriber)
	}
	return nil
}

type queryable struct {
	objdb object.DB
	refdb ref.DB
}

func (q queryable) Matching(ctx context.Context, m retro.Matcher) (retro.MatcherResults, error) {

	// Resolve the head ref for the given ctx
	headRef, err := q.refdb.Retrieve(refFromCtx(ctx))
	if err != nil {
		return nil, errors.Wrapf(err, "unknown ref %s", refFromCtx(ctx))
	}

	err = q.gatherMatches(ctx, m, headRef)

	return nil, err
}

// TODO: This can/should be modified to honor the possible parallel histories
// in case any checkpoint has two parents, in which case the histories should
// probably be interleaved.
//
//
// 1. gather matches takes a single hash, retrieves the checkpoint and affix, and runs
//    them all through the matcher
//
//
// TODO: This could easily run the matcher concurrently across all events and the affix and
// the checkpoint. Write a benchmark suite before testing this.
//
func (q queryable) gatherMatches(ctx context.Context, m retro.Matcher, cpHash retro.Hash) error {
	cp, err := q.retrieveCheckpoint(cpHash)
	if err != nil {
		return errors.Wrap(err, "can't gather matches, error retreiving checkpoint")
	}

	af, err := q.retrieveAffix(cp.AffixHash)
	if err != nil {
		return errors.Wrap(err, "can't gather matches, error retrieving affix")
	}

	for _, ph := range cp.ParentHashes {
		q.gatherMatches(ctx, m, ph)
	}

	var (
		matcherErrs chan error
	)

	_ = matcherErrs

	go q.doesCheckpointMatch(ctx, m, *cp)
	go q.doesAffixMatch(ctx, m, *af)

	return nil
}

func (q queryable) retrieveCheckpoint(h retro.Hash) (*packing.Checkpoint, error) {

	var jp *packing.JSONPacker

	var (
		checkpoint packing.Checkpoint
		err        error
	)

	packedCheckpoint, err := q.objdb.RetrievePacked(h.String())
	if err != nil {
		// TODO: test this case
		return nil, errors.Wrap(err, fmt.Sprintf("can't read object %s", h.String()))
	}

	if packedCheckpoint.Type() != packing.ObjectTypeCheckpoint {
		// TODO: test this case
		return nil, errors.Wrap(err, fmt.Sprintf("object was not a %s but a %s", packing.ObjectTypeCheckpoint, packedCheckpoint.Type()))
	}

	checkpoint, err = jp.UnpackCheckpoint(packedCheckpoint.Contents())
	if err != nil {
		// TODO: test this case
		return nil, errors.Wrap(err, fmt.Sprintf("can't read object %s", h.String()))
	}

	return &checkpoint, err
}

func (q queryable) retrieveAffix(h retro.Hash) (*packing.Affix, error) {

	var jp *packing.JSONPacker

	var (
		affix packing.Affix
		err   error
	)

	// Unpack the Affix
	packedAffix, err := q.objdb.RetrievePacked(h.String())
	if err != nil {
		// TODO: test this case
		return nil, errors.Wrap(err, fmt.Sprintf("retrieve affix %s", h.String()))
	}

	if packedAffix.Type() != packing.ObjectTypeAffix {
		// TODO: test this case
		return nil, errors.Wrap(err, fmt.Sprintf("object was not a %s but a %s", packing.ObjectTypeAffix, packedAffix.Type()))
	}

	affix, err = jp.UnpackAffix(packedAffix.Contents())
	if err != nil {
		// TODO: test this case
		return nil, errors.Wrap(err, fmt.Sprintf("unpack affix %s s", h.String()))
	}

	return &affix, nil
}

func (q queryable) doesCheckpointMatch(_ context.Context, m retro.Matcher, cp packing.Checkpoint) (matcher.Result, error) {
	return m.DoesMatch(cp)
}

func (q queryable) doesAffixMatch(_ context.Context, m retro.Matcher, af packing.Affix) (matcher.Result, error) {
	return m.DoesMatch(af)
}
