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

	"github.com/golang-collections/collections/stack"
	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/ref"
	"github.com/retro-framework/go-retro/framework/storage"
	"github.com/retro-framework/go-retro/framework/storage/memory"
	"github.com/retro-framework/go-retro/framework/types"
)

// DefaultBranchName is defined so that without an override changes
// will move the ref named by this branch name
const DefaultBranchName = "refs/heads/master"

// NewSimpleStub returns a simple Depot stub which will yield the given events in the fixture
// as a single checkpoint with a single affix with a generic set of placeholder metadata.
func NewSimpleStub(t *testing.T,
	objDB object.DB,
	refDB ref.DB,
	fixture map[string][]types.EventNameTuple,
) types.Depot {
	var (
		jp    = packing.NewJSONPacker()
		affix = packing.Affix{}
	)
	for aggName, evNameTuples := range fixture {
		var (
			evHashesForAffix []types.Hash
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
		affix[types.PartitionName(aggName)] = evHashesForAffix
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

func NewSimple(odb object.DB, refdb ref.DB) types.Depot {
	return &Simple{objdb: odb, refdb: refdb}
}

// EmptySimpleMemory returns an empty depot to keep the type system happy
func EmptySimpleMemory() types.Depot {
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
func (s *Simple) HeadPointer(ctx context.Context) (types.Hash, error) {
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
	} else {
		fmt.Print("couldn't upgrade store")
	}
	if lrefdb, ok := s.refdb.(ref.ListableStore); ok {
		m, err := lrefdb.Ls()
		if err != nil {
			fmt.Fprintf(w, "err: %s", err)
		}
		for k, v := range m {
			fmt.Fprintf(w, "%s -> %s\n", k, v)
		}
	} else {
		fmt.Print("couldn't upgrade store")
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

	subscribers []chan<- types.RefMove
}

// TODO: make this respect the actual value that might come in a context
func refFromCtx(ctx context.Context) string {
	return DefaultBranchName
}

func (s Simple) Claim(ctx context.Context, partition string) bool {
	// TODO: Implement locking properly
	return true
}

func (s Simple) Release(partition string) {
	// TODO: Implement locking properly
	return
}

func (s Simple) Exists(partitionName types.PartitionName) bool {
	found, _ := simplePartitionExistenceChecker{
		objdb:   s.objdb,
		refdb:   s.refdb,
		pattern: partitionName,
		matcher: GlobPatternMatcher{},
	}.Exists(context.TODO(), partitionName)
	return found
}

// Rehydrate replays the events onto an aggregate, it's kinda brutal in that it completely
// walks from the tip until the first orphan checkpoint, and stacks all relevant partitions
// then emits them all, it's very expensive.
func (s Simple) Rehydrate(ctx context.Context, dst types.Aggregate, partitionName types.PartitionName) error {
	return simpleAggregateRehydrater{
		objdb:   s.objdb,
		refdb:   s.refdb,
		pattern: partitionName,
		matcher: GlobPatternMatcher{},
	}.Rehydrate(ctx, dst, partitionName)
}

// Glob makes the world go round
func (s *Simple) Glob(_ context.Context, partition string) types.PartitionIterator {
	var subscriberNotificationCh = make(chan types.RefMove)
	s.subscribers = append(s.subscribers, subscriberNotificationCh)
	return &simplePartitionIterator{
		objdb:          s.objdb,
		refdb:          s.refdb,
		pattern:        partition,
		matcher:        GlobPatternMatcher{},
		subscribedOn:   subscriberNotificationCh,
		eventIterators: make(map[string]*simpleEventIterator),
	}
}

// StorePacked takes a variable number of hashed objects, packs and stores them
// in the object store backing the Simple Depot
func (s Simple) StorePacked(packed ...types.HashedObject) error {
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
func (s Simple) MoveHeadPointer(old, new types.Hash) error {
	_, err := s.refdb.Write(DefaultBranchName, new)
	if err == nil {
		s.notifySubscribers(old, new)
	}
	return err
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
func (s Simple) notifySubscribers(old, new types.Hash) error {
	for _, subscriber := range s.subscribers {
		go func(subscriber chan<- types.RefMove) {
			select {
			case subscriber <- types.RefMove{Old: old, New: new}:
				// TODO: something about metrics ?
			case <-time.After(1 * time.Minute):
				fmt.Fprintf(os.Stderr, "blocked for one minute waiting to notify subscriber, skipping.")
			}
		}(subscriber)
	}
	return nil
}

type relevantCheckpoint struct {
	time           time.Time
	checkpointHash types.Hash
	affix          packing.Affix
}

func (rc relevantCheckpoint) String() string {
	return fmt.Sprintf("Relevant Checkpoint: %s", rc.checkpointHash.String())
}

type cpAffixStack struct {
	s stack.Stack

	knownPartitions []types.PartitionName
}

// Push pushes a relavantCheckpoint onto a stack as we walk
// the object graph. It also maintains a youngest-to-oldest
func (os *cpAffixStack) Push(rc relevantCheckpoint) {
	os.s.Push(rc)
	for partitionName := range rc.affix {
		var partitionNameKnown bool
		for _, kp := range os.knownPartitions {
			if partitionName == kp {
				partitionNameKnown = true
			}
		}
		if !partitionNameKnown {
			os.knownPartitions = append(os.knownPartitions, partitionName)
		}
	}
}

func (os *cpAffixStack) Pop() *relevantCheckpoint {
	v := os.s.Pop()
	if v == nil {
		return nil
	}
	rcp := v.(relevantCheckpoint)
	return &rcp
}

// PatternMatcher defines a single function interface
// for matching patterns. It is used to compare the aggregate
// paths within an affix to the aggregate name being searched
// for. In a sane implementation it should support at least
// POSIX globbing and perhaps even Regular Expressions to
// allow for matching such as `users/*` or similar.
//
// In testing, this pattern matcher may be replaced with a
// no-op or static matcher.
type PatternMatcher interface {
	DoesMatch(pattern, partition string) (bool, error)
}
