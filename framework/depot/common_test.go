// +build integration

package depot

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/retro-framework/go-retro/events" // TODO: Fix this don't reach out of framework!
	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/ref"
	"github.com/retro-framework/go-retro/framework/retro"
	"github.com/retro-framework/go-retro/framework/storage/fs"
	"github.com/retro-framework/go-retro/framework/storage/memory"
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

	// Events
	var (
		// common fixtures
		setAuthorName1, _          = jp.PackEvent("set_author_name", DummyEvSetAuthorName{"Maxine Mustermann"})
		setArticleTitle1, _        = jp.PackEvent("set_article_title", DummyEvSetArticleTitle{"event graph for noobs"})
		associateArticleAuthor1, _ = jp.PackEvent("associate_article_author", DummyEvAssociateArticleAuthor{"author/maxine"})
		setArticleTitle2, _        = jp.PackEvent("set_article_title", DummyEvSetArticleTitle{"learning event graph"})
		setArticleBody1, _         = jp.PackEvent("set_article_body", DummyEvSetArticleBody{"lorem ipsum ..."})

		// extended fixtures
		setAuthorName2, _          = jp.PackEvent("set_author_name", DummyEvSetAuthorName{"Paul Peterson"})
		associateArticleAuthor2, _ = jp.PackEvent("associate_article_author", DummyEvAssociateArticleAuthor{"author/paul"})
	)

	// Affixes
	var (
		affixOne, _   = jp.PackAffix(packing.Affix{"author/maxine": []retro.Hash{setAuthorName1.Hash()}})
		affixTwo, _   = jp.PackAffix(packing.Affix{"article/first": []retro.Hash{setArticleTitle1.Hash(), associateArticleAuthor1.Hash()}})
		affixThree, _ = jp.PackAffix(packing.Affix{"article/first": []retro.Hash{setArticleTitle2.Hash(), setArticleBody1.Hash()}})

		// extended
		affixFourA, _ = jp.PackAffix(packing.Affix{
			"author/paul":    []retro.Hash{setAuthorName2.Hash()},
			"article/second": []retro.Hash{associateArticleAuthor2.Hash()},
		})

		affixFourB, _ = jp.PackAffix(packing.Affix{
			"article/first": []retro.Hash{associateArticleAuthor2.Hash()},
		})
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
			ParentHashes: []retro.Hash{checkpointOne.Hash()},
		})

		checkpointThree, _ = jp.PackCheckpoint(packing.Checkpoint{
			AffixHash:   affixThree.Hash(),
			CommandDesc: []byte(`{"update":"article"}`),
			Fields: map[string]string{
				"session": "hello world",
				"date":    clock.Now().Format(time.RFC3339),
			},
			ParentHashes: []retro.Hash{checkpointTwo.Hash()},
		})

		// Extend
		checkpointFourA, _ = jp.PackCheckpoint(packing.Checkpoint{
			AffixHash:   affixFourA.Hash(),
			CommandDesc: []byte(`{"update":"article"}`),
			Fields: map[string]string{
				"session": "hello world",
				"date":    clock.Now().Format(time.RFC3339),
			},
			ParentHashes: []retro.Hash{checkpointThree.Hash()},
		})

		checkpointFourB, _ = jp.PackCheckpoint(packing.Checkpoint{
			AffixHash:   affixFourB.Hash(),
			CommandDesc: []byte(`{"update":"article"}`),
			Fields: map[string]string{
				"session": "hello world",
				"date":    clock.Now().Format(time.RFC3339),
			},
			ParentHashes: []retro.Hash{checkpointThree.Hash()},
		})
	)

	baseTmpdir, err := ioutil.TempDir("", "depot_common_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(baseTmpdir)

	// EventManifest is used to map event names back to prototypes
	// of the original so that we can reconstruct them.
	var evManifest = events.NewManifest()
	evManifest.RegisterAs("set_author_name", &DummyEvSetAuthorName{})
	evManifest.RegisterAs("set_article_title", &DummyEvSetArticleTitle{})
	evManifest.RegisterAs("set_article_body", &DummyEvSetArticleBody{})
	evManifest.RegisterAs("associate_article_author", &DummyEvAssociateArticleAuthor{})

	odbs := map[string]func() object.DB{
		"memory": func() object.DB {
			return &memory.ObjectStore{}
		},
		"fs": func() object.DB {
			// This dir is within baseTmpdir and will be removed
			// when Test_Depot() ends
			dir, err := ioutil.TempDir(baseTmpdir, "odb")
			if err != nil {
				t.Fatal(err)
			}
			return &fs.ObjectStore{BasePath: dir}
		},
	}
	refdbs := map[string]func() ref.DB{
		"memory": func() ref.DB {
			return &memory.RefStore{}
		},
		"fs": func() ref.DB {
			// This dir is within baseTmpdir and will be removed
			// when Test_Depot() ends
			dir, err := ioutil.TempDir(baseTmpdir, "refdb")
			if err != nil {
				t.Fatal(err)
			}
			return &fs.RefStore{BasePath: dir}
		},
	}

	var populateOdb = func(odb object.DB) object.DB {
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
		return odb
	}

	depotFns := map[string]func() retro.Depot{
		"memory": func() retro.Depot {
			var odb = populateOdb(odbs["memory"]())
			return &Simple{
				objdb: odb,
				refdb: refdbs["memory"](),
			}
		},
		// "fs":        &Simple{objdb: odbs["fs"], refdb: refdbs["fs"], eventManifest: evManifest},
		// "fs+memory": &Simple{objdb: odbs["memory"], refdb: refdbs["fs"], eventManifest: evManifest},
		// "memory+fs": &Simple{objdb: odbs["fs"], refdb: refdbs["memory"], eventManifest: evManifest},
	}

	for name, depotFn := range depotFns {

		t.Run(name, func(t *testing.T) {
			t.Run("correctly checking for existence of aggregates", func(t *testing.T) {
				t.Skip("not implemented yet")
			})
		})

		t.Run(name, func(t *testing.T) {

			t.Run("iterates over correct events in correct order", func(t *testing.T) {

				var depot = depotFn()
				depot.MoveHeadPointer(nil, checkpointThree.Hash())

				var (
					expectedResult = map[retro.PartitionName][]string{
						"author/maxine": []string{"set_author_name"},
						"article/first": []string{
							"set_article_title",
							"associate_article_author",
							"set_article_title",
							"set_article_body",
						},
					}

					errs  = make(chan error, 1)
					diffs = make(chan string)

					foundEvs = make(chan struct {
						pn  retro.PartitionName
						pEv retro.PersistedEvent
					})
				)

				var ctx, cancelFn = context.WithTimeout(context.Background(), 1*time.Second)
				defer cancelFn()

				var (
					// start conditions, we're globbing for any event on any partition
					partitionInterator          = depot.Watch(ctx, "*")
					partitions, partitionErrors = partitionInterator.Partitions(ctx)
				)

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
					pn  retro.PartitionName
					pEv retro.PersistedEvent
				}) {
					var seenResults = make(map[retro.PartitionName][]string)
					for recv := range received {
						// TODO: redundant, I think
						// if _, ok := seenResults[recv.pn]; !ok {
						// 	seenResults[recv.pn] = make([]string. 0)
						// }
						seenResults[recv.pn] = append(seenResults[recv.pn], recv.pEv.Name())
						diffs <- cmp.Diff(expectedResult, seenResults)
					}
				}(ctx, foundEvs)

				// Event handler makes a tuple of the data about the event, and sends
				// it on the channel where the results are being collected
				var eventHandler = func(ctx context.Context, pn retro.PartitionName, pEv retro.PersistedEvent) {
					foundEvs <- struct {
						pn  retro.PartitionName
						pEv retro.PersistedEvent
					}{
						pn:  pn,
						pEv: pEv,
					}
				}

				go func() {
					for partition := range partitions {
						go func(ctx context.Context, evIter retro.EventIterator) {
							events, _ := evIter.Events(ctx)
							for event := range events {
								eventHandler(ctx, retro.PartitionName(evIter.Pattern()), event)
							}
						}(ctx, partition)
					}
				}()

				var lastDiff string
				for {
					select {
					case err := <-errs:
						if err != nil {
							t.Fatal(err)
						}
					case lastDiff = <-diffs:
						if lastDiff == "" {
							return
						}
					case <-ctx.Done():
						t.Errorf("\nexpectedResults, seenResults differs: (-want +got)\n%s", lastDiff)
						t.Fatal("test failed", ctx.Err())
					}
				}
			})

			t.Run("propagates new partitions after a consumer has consumed all that existed at start time", func(t *testing.T) {

				var depot = depotFn()
				depot.MoveHeadPointer(nil, checkpointThree.Hash())

				var ctx, cancelFn = context.WithTimeout(context.Background(), 1*time.Second)
				defer cancelFn()

				var seenExpected int
				var partitionInterator = depot.Watch(ctx, "*")
				var partitions, partitionErrors = partitionInterator.Partitions(ctx)

				// Test fixture has two existing partitions consume them both
				_ = <-partitions
				_ = <-partitions

				//
				// In a goroutine we will add a new partition and move the head pointer
				// and the subscription mechanism should kick in, we could also do this
				// on the main thread and do the test in a goroutine, it should make
				// no difference
				//
				go func() {
					depot.StorePacked(setAuthorName2)
					depot.StorePacked(associateArticleAuthor2)
					depot.StorePacked(affixFourA)
					depot.StorePacked(checkpointFourA)
					depot.MoveHeadPointer(checkpointThree.Hash(), checkpointFourA.Hash())
				}()

				for {
					select {
					case <-ctx.Done():
						t.Fatal(ctx.Err(), "waiting for expected condition")
					case err, ok := <-partitionErrors:
						if ok {
							t.Fatal(err)
						}
					case newPartition, ok := <-partitions:
						if ok && (newPartition.Pattern() == "author/paul" || newPartition.Pattern() == "article/second") {
							seenExpected += 1
						} else {
							t.Errorf("unexpected partition seen %s", newPartition.Pattern())
						}
						if seenExpected == 2 {
							return // Quietly return as soon as we see two extra ones.
						}
					}
				}
			})

			t.Run("propagates new events after a consumer has reached the head pointer", func(t *testing.T) {
				var depot = depotFn()
				depot.MoveHeadPointer(nil, checkpointThree.Hash())

				var ctx, cancelFn = context.WithTimeout(context.Background(), 1*time.Second)
				defer cancelFn()

				var success = make(chan bool)
				var partitionInterator = depot.Watch(ctx, "*")
				var partitions, partitionErrors = partitionInterator.Partitions(ctx)

				var saveNewDataAndMoveHeadPointer = func() {
					depot.StorePacked(associateArticleAuthor2)
					depot.StorePacked(affixFourB)
					depot.StorePacked(checkpointFourB)
					depot.MoveHeadPointer(checkpointThree.Hash(), checkpointFourB.Hash())
				}

				var handleEvents = func(ctx context.Context, evi retro.EventIterator) {
					events, errors := evi.Events(ctx)
					for {
						select {

						case e, ok := <-events:
							if !ok {
								events = nil
								return
							}
							// This code path consumes _all_ events not only after the head
							// pointer move. The guard in the for{ select {}} } loop below
							// however ensures that the head pointer is not moved until we've
							// consumed at least two partitions of events. This means that this
							// test should testing the right thing, but it would benefit from
							// a refactoring.
							if e.Name() == "associate_article_author" &&
								bytes.Compare(e.Bytes(), []byte(`{"AuthorURN":"author/maxine"}`)) == 0 {
								success <- true
								return
							}
						case err, ok := <-errors:
							if !ok {
								errors = nil
								return
							}
							if err != nil {
								t.Error("event emitter erred", err)
							}
						case <-ctx.Done():
							return
						}
					}
				}

				var seenPartitions int
				for {
					select {
					case p, ok := <-partitions:
						if ok {
							go handleEvents(ctx, p)
							seenPartitions++
							if seenPartitions == 2 {
								saveNewDataAndMoveHeadPointer()
							}
						}
					case pErr := <-partitionErrors:
						t.Error("partiton iterator emitted error", pErr)
					case <-success:
						ctx = nil // prevent reading from ctx
						return
					case <-ctx.Done():
						t.Fatal("test failed", ctx.Err())
					}
				}

			})

			t.Run("has a simple next API that does not rely on channels", func(t *testing.T) {

				var success = make(chan bool)

				var depot = depotFn()
				depot.MoveHeadPointer(nil, checkpointThree.Hash())

				var ctx, cancelFn = context.WithTimeout(context.Background(), 1*time.Second)
				defer cancelFn()

				go func() {
					var authors = depot.Watch(ctx, "author/*")
					for {
						authorEvents, err := authors.Next(ctx)
						if err == Done {
							continue
						}
						if err != nil {
							t.Fatal(err)
							return
						}
						if authorEvents != nil && authorEvents.Pattern() == "author/maxine" {
							for {
								var ev, err = authorEvents.Next(ctx)
								if err == Done {
									continue
								}
								if err != nil {
									t.Fatal(err)
									return
								}
								// TODO: without the nil guard the line below sometimes
								// sees a nil event, the two conditions above should
								// have prevented that.
								if ev != nil && ev.Name() == "set_author_name" {
									success <- true
								}
							}
						}
					}
				}()

				select {
				case <-ctx.Done():
					t.Fatal("test failed", ctx.Err())
				case <-success:
					return
				}
			})
		})
	}
}
