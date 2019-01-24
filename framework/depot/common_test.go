package depot

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/ref"
	"github.com/retro-framework/go-retro/framework/storage/fs"
	"github.com/retro-framework/go-retro/framework/storage/memory"
	"github.com/retro-framework/go-retro/framework/types"
)

type DummyEvSetAuthorName struct {
	Name string
}

type DummyEvSetArticleTitle struct {
	Title string
}

type DummyEvSetArticleBody struct {
	Name string
}

type DummyEvAssociateArticleAuthor struct {
	AuthorURN string
}

func Example() {
	tmpdir, err := ioutil.TempDir("", "example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	_ = Simple{
		objdb: &fs.ObjectStore{BasePath: tmpdir},
		refdb: &fs.RefStore{BasePath: tmpdir},
	}
}

type Predictable5sJumpClock struct {
	t     time.Time
	calls int
}

func (c *Predictable5sJumpClock) Now() time.Time {
	var next = c.t.Add(time.Duration((5 * c.calls)) * time.Second)
	c.calls = c.calls + 1
	return next
}

func Test_Depot(t *testing.T) {

	var jp = packing.NewJSONPacker()

	// var expectedEvNames = []string{"set_author_name", "set_article_title","associate_article_author","set_article_title","set_article_body"}

	// Events
	var (
		setAuthorName1, _          = jp.PackEvent("set_author_name", DummyEvSetAuthorName{"Maxine Mustermann"})
		setArticleTitle1, _        = jp.PackEvent("set_article_title", DummyEvSetArticleTitle{"event graph for noobs"})
		associateArticleAuthor1, _ = jp.PackEvent("associate_article_author", DummyEvAssociateArticleAuthor{"author/maxine"})
		setArticleTitle2, _        = jp.PackEvent("set_article_title", DummyEvSetArticleTitle{"learning event graph"})
		setArticleBody1, _         = jp.PackEvent("set_article_body", DummyEvSetArticleBody{"lorem ipsum ..."})
	)

	// Affixes
	var (
		affixOne, _   = jp.PackAffix(packing.Affix{"author/maxine": []types.Hash{setAuthorName1.Hash()}})
		affixTwo, _   = jp.PackAffix(packing.Affix{"article/first": []types.Hash{setArticleTitle1.Hash(), associateArticleAuthor1.Hash()}})
		affixThree, _ = jp.PackAffix(packing.Affix{"article/first": []types.Hash{setArticleTitle2.Hash(), setArticleBody1.Hash()}})
	)

	var clock = Predictable5sJumpClock{}

	// Checkpoints
	var (
		checkpointOne, _ = jp.PackCheckpoint(packing.Checkpoint{
			AffixHash:   affixOne.Hash(),
			CommandDesc: []byte(`{"create":"author"}`),
			Fields: map[string]string{
				"session": "hello world",
				"date":    clock.Now().Format(time.RFC3339),
			},
		})

		checkpointTwo, _ = jp.PackCheckpoint(packing.Checkpoint{
			AffixHash:   affixTwo.Hash(),
			CommandDesc: []byte(`{"draft":"article"}`),
			Fields: map[string]string{
				"session": "hello world",
				"date":    clock.Now().Format(time.RFC3339),
			},
			ParentHashes: []types.Hash{checkpointOne.Hash()},
		})

		checkpointThree, _ = jp.PackCheckpoint(packing.Checkpoint{
			AffixHash:   affixThree.Hash(),
			CommandDesc: []byte(`{"update":"article"}`),
			Fields: map[string]string{
				"session": "hello world",
				"date":    clock.Now().Format(time.RFC3339),
			},
			ParentHashes: []types.Hash{checkpointTwo.Hash()},
		})
	)

	tmpdir, err := ioutil.TempDir("", "depot_common_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	// EventManifest is used to map event names back to prototypes
	// of the original so that we can reconstruct them.
	var evManifest = events.NewManifest()
	evManifest.RegisterAs("set_author_name", &DummyEvSetAuthorName{})
	evManifest.RegisterAs("set_article_title", &DummyEvSetArticleTitle{})
	evManifest.RegisterAs("set_article_body", &DummyEvSetArticleBody{})
	evManifest.RegisterAs("associate_article_author", &DummyEvAssociateArticleAuthor{})

	odbs := map[string]object.DB{
		"memory": &memory.ObjectStore{},
		"fs":     &fs.ObjectStore{BasePath: tmpdir},
	}
	refdbs := map[string]ref.DB{
		"memory": &memory.RefStore{},
		"fs":     &fs.RefStore{BasePath: tmpdir},
	}
	depots := map[string]types.Depot{
		"memory": Simple{objdb: odbs["memory"], refdb: refdbs["memory"], eventManifest: evManifest},
		// "fs":     Simple{objdb: odbs["fs"], refdb: refdbs["fs"]},
		// "fs+memory": Simple{objdb: odbs["memory"], refdb: refdbs["fs"]},
		// "memory+fs": Simple{objdb: odbs["fs"], refdb: refdbs["memory"]},
	}

	for _, odb := range odbs {
		odb.WritePacked(setAuthorName1)
		odb.WritePacked(setArticleTitle1)
		odb.WritePacked(associateArticleAuthor1)
		odb.WritePacked(setArticleTitle2)
		odb.WritePacked(setArticleBody1)

		odb.WritePacked(affixOne)
		odb.WritePacked(affixTwo)
		odb.WritePacked(affixThree)

		odb.WritePacked(checkpointOne)
		odb.WritePacked(checkpointTwo)
		odb.WritePacked(checkpointThree)
	}

	for _, refdb := range refdbs {
		refdb.Write(DefaultBranchName, checkpointThree.Hash())
	}

	for name, depot := range depots {

		t.Run(name, func(t *testing.T) {
			t.Run("correctly checking for existence of aggregates", func(t *testing.T) {
				t.Skip("not implemented yet")
			})
		})

		t.Run(name, func(t *testing.T) {

			t.Run("iterates over correct events in correct order", func(t *testing.T) {

				var (
					expectedResult = map[types.PartitionName][]types.EventNameTuple{
						"author/maxine": []types.EventNameTuple{
							{Name: "set_author_name", Event: DummyEvSetAuthorName{"Maxine Mustermann"}},
						},
						"article/first": []types.EventNameTuple{
							{Name: "set_article_title", Event: DummyEvSetArticleTitle{"event graph for noobs"}},
							{Name: "associate_article_author", Event: DummyEvAssociateArticleAuthor{"author/maxine"}},
							{Name: "set_article_title", Event: DummyEvSetArticleTitle{"learning event graph"}},
							{Name: "set_article_body", Event: DummyEvSetArticleBody{"lorem ipsum ..."}},
						},
					}

					errs = make(chan error, 1)

					expectedConditionSeen = false
					conditionMutex        = &sync.RWMutex{}
					lastDiff              string
					lastDiffLock          sync.RWMutex

					foundEvs = make(chan struct {
						pn  types.PartitionName
						pEv types.PersistedEvent
						ev  types.Event
					})

					// cancelFn will ensure we always clean up, this is what
					// we always use to proportage the exit condition
					ctx, cancelFn = context.WithTimeout(context.Background(), 5*time.Second)

					// start conditions, we're globbing for any event on any partition
					partitionInterator          = depot.Glob(ctx, "*")
					partitions, partitionErrors = partitionInterator.Partitions(ctx)
				)

				defer cancelFn()

				// This go routine aborts the test when the context timeout
				// is reached. We can't FatalF in a go-routine, that kills
				// onyl that goroutine. Send an error to the main goroutine.
				go func(ctx context.Context) {
					<-ctx.Done()
					conditionMutex.Lock()
					defer conditionMutex.Unlock()
					if !expectedConditionSeen {
						lastDiffLock.RLock()
						errs <- fmt.Errorf("timeout reached, failing, difference: %s", lastDiff)
						lastDiffLock.RUnlock()
					}
				}(ctx)

				// This go routine handles the case that we found matcher errors
				// in either the partition iterator or the event iterator.
				go func(partitionErrs <-chan error) {
					partitionErr := <-partitionErrs
					// we should see a nil through this channel every time
					// we don't fail, so squash those, as <-errs will finalize
					// the test whatever value we send.
					if err != nil {
						errs <- partitionErr
					}

				}(partitionErrors)

				// This Go routine is waiting for tuples with a partition name
				// and an event, they are assumed to arrive in chronological order.
				// This goroutine passes the test by closing the other iterators
				// when it has all the information it expected to see.
				go func(ctx context.Context, received chan struct {
					pn  types.PartitionName
					pEv types.PersistedEvent
					ev  types.Event
				}) {

					var lockResults sync.Mutex
					var seenResults = make(map[types.PartitionName][]types.EventNameTuple)

					for recv := range received {

						lockResults.Lock()
						if _, ok := seenResults[recv.pn]; !ok {
							seenResults[recv.pn] = []types.EventNameTuple{}
						}
						seenResults[recv.pn] = append(seenResults[recv.pn], types.EventNameTuple{Name: recv.pEv.Name(), Event: recv.ev})
						lockResults.Unlock()

						lastDiffLock.Lock()
						lastDiff = cmp.Diff(expectedResult, seenResults)

						if lastDiff == "" {
							errs <- nil // signal the end of the test
							lastDiffLock.Unlock()
							return
						}
						lastDiffLock.Unlock()

					}
				}(ctx, foundEvs)

				// Event handler makes a tuple of the data about the event, and sends
				// it on the channel where the results are being collected
				var eventHandler = func(ctx context.Context, pn types.PartitionName, pEv types.PersistedEvent) {
					ev, err := pEv.Event()
					if err != nil {
						errs <- errors.Wrap(err, "error in event handler")
					}
					foundEvs <- struct {
						pn  types.PartitionName
						pEv types.PersistedEvent
						ev  types.Event
					}{
						pn:  pn,
						pEv: pEv,
						ev:  ev,
					}
				}

				var partitionHandler = func(ctx context.Context, evIter types.EventIterator) {
					events, _ := evIter.Events(ctx)
					for event := range events {
						go eventHandler(ctx, types.PartitionName(evIter.Pattern()), event)
					}
				}

				for partition := range partitions {
					go partitionHandler(ctx, partition)
				}

				// wait for it to signal, anything other than nil
				// indicates an error. whoever sends that error
				// should send a useful debugging message.
				err := <-errs
				if err != nil {
					t.Fatal(err)
				}

			})
		})
	}
}
