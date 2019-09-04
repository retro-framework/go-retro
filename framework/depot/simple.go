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
	"github.com/retro-framework/go-retro/framework/ctxkey"
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
// the value of the current head pointer from the refdb
// it uses the context to try and get a branch name, and
// in case of failure falls back to the default branch
// name.
func (s *Simple) HeadPointer(ctx context.Context) (retro.Hash, error) {
	ptr, err := s.refdb.Retrieve(ctxkey.Ref(ctx))
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
//
// TODO: make sure that pointer is currently at OLD if provided
// before moving it.
func (s Simple) MoveHeadPointer(ctx context.Context, old, new retro.Hash) error {
	_, err := s.refdb.Write(ctxkey.Ref(ctx), new)
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

// type checkpointAtTime struct {
// 	t         time.Time
// 	cpHashStr string
// 	// TODO: include matcher.Result here?
// }

// type MatcherInstance struct {
// 	q  queryable
// 	wg sync.WaitGroup
// }

func (q queryable) Matching(ctx context.Context, m retro.Matcher) (retro.MatcherResults, error) {

	// 1. Find the get a chronological index
	//    for the current branch.
	//
	// TODO: subscribers don't really respect
	//       branch names yet, so this isn't
	//       really reliable. (see the data structure
	//       for head pointer moves, it is not branch
	//       aware)
	// ci, _ := index.ChronologicalForBranch(ctxkey.Ref(ctx), q.objdb, q.refdb)

	// 2. Register that index as a subscriber to receive updates
	//    about head pointer movements

	// 3. Consume that resource indefinitely passing the checkpoints
	//    to the matcher

	// 	// Resolve the head ref for the given ctx

	// 	fmt.Println("gathering matches")

	// 	var (
	// 		results    = matcher.Results{}
	// 		matchedCPs = make(chan checkpointAtTime)
	// 		// matcherErrs = make(chan error)

	// 		finished      = make(chan struct{})
	// 		noMoreResults = make(chan struct{})
	// 	)

	// 	go q.gatherMatches(ctx, matchedCPs, noMoreResults, m, headRef)

	// 	go func(rCh <-chan checkpointAtTime, noMoreResults <-chan struct{}) {
	// 		for {
	// 			select {
	// 			case cpAtTime := <-rCh:
	// 				fmt.Println("Matched a checkpoint")
	// 				results = results.Insert(cpAtTime.t, cpAtTime.cpHashStr)
	// 			case <-noMoreResults:
	// 				fmt.Println("signalling EOL, no more checkpoints")
	// 				finished <- struct{}{}
	// 			}
	// 		}
	// 		fmt.Printf("Results all found are %#v\n", results)
	// 	}(matchedCPs, noMoreResults)

	// 	fmt.Println("waiting until we find only checkpoints with no ancestors")
	// 	fmt.Println("done")

	return nil, nil
}

// func (q queryable) gatherMatches(ctx context.Context, cpAtTimeCh chan<- checkpointAtTime, noMoreResultsCh chan<- struct{}, m retro.Matcher, cpHash retro.Hash) error {

// 	fmt.Println("checkpoints +1")

// 	fmt.Println("in gather matches with", cpHash)

// 	cp, err := q.retrieveCheckpoint(cpHash)
// 	if err != nil {
// 		return errors.Wrap(err, "can't gather matches, error retreiving checkpoint")
// 	}

// 	af, err := q.retrieveAffix(cp.AffixHash)
// 	if err != nil {
// 		return errors.Wrap(err, "can't gather matches, error retrieving affix")
// 	}

// 	for _, b := range *af {
// 		for _, ev := range b {
// 			fmt.Println("EvHash", ev)
// 		}
// 	}

// 	for _, ph := range cp.ParentHashes {

// 		go q.gatherMatches(ctx, cpAtTimeCh, noMoreResultsCh, m, ph)
// 	}

// 	// if we get an early match we can avoid maybe
// 	// looking up all the contents of the db here
// 	// e.g if matching on a session id, no need to look
// 	// at the affix and events

// 	var timeFromCP = func(cp *packing.Checkpoint) time.Time {
// 		t, err := time.Parse(time.RFC3339, (*cp).Fields["date"])
// 		if err != nil {
// 			fmt.Println("extremely serious error, could not parse a time we previously validated", err)
// 		}
// 		return t
// 	}

// 	if match, _ := m.DoesMatch(*cp); match.Success() {
// 		cpAtTimeCh <- checkpointAtTime{t: timeFromCP(cp), cpHashStr: cpHash.String()}
// 	}

// 	if match, _ := m.DoesMatch(*af); match.Success() {
// 		fmt.Println("affix matched")
// 	}

// 	time.Sleep(4)

// 	return nil
// }

// func (q queryable) retrieveAffix(h retro.Hash) (*packing.Affix, error) {

// 	var jp *packing.JSONPacker

// 	var (
// 		affix packing.Affix
// 		err   error
// 	)

// 	// Unpack the Affix
// 	packedAffix, err := q.objdb.RetrievePacked(h.String())
// 	if err != nil {
// 		// TODO: test this case
// 		return nil, errors.Wrap(err, fmt.Sprintf("retrieve affix %s", h.String()))
// 	}

// 	if packedAffix.Type() != packing.ObjectTypeAffix {
// 		// TODO: test this case
// 		return nil, errors.Wrap(err, fmt.Sprintf("object was not a %s but a %s", packing.ObjectTypeAffix, packedAffix.Type()))
// 	}

// 	affix, err = jp.UnpackAffix(packedAffix.Contents())
// 	if err != nil {
// 		// TODO: test this case
// 		return nil, errors.Wrap(err, fmt.Sprintf("unpack affix %s s", h.String()))
// 	}

// 	return &affix, nil
// }

// // func (q queryable) doesCheckpointMatch(_ context.Context, m retro.Matcher, cp packing.Checkpoint) (matcher.Result, error) {
// // 	return m.DoesMatch(cp)
// // }

// // func (q queryable) doesAffixMatch(_ context.Context, m retro.Matcher, af packing.Affix) (matcher.Result, error) {
// // 	return m.DoesMatch(af)
// // }
