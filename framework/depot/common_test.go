package depot

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	// "github.com/fortytw2/leaktest"
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

func Test_Depot(t *testing.T) {

	// defer leaktest.Check(t)()

	var jp = packing.NewJSONPacker()

	// Events
	var (
		setAuthorName1, _ = jp.PackEvent("set_author_name", DummyEvSetAuthorName{"Maxine Mustermann"})

		setArticleTitle1, _        = jp.PackEvent("set_article_title", DummyEvSetArticleTitle{"event graph for noobs"})
		associateArticleAuthor1, _ = jp.PackEvent("associate_article_author", DummyEvAssociateArticleAuthor{"author/maxine"})

		setArticleTitle2, _ = jp.PackEvent("set_article_title", DummyEvSetArticleTitle{"learning event graph"})
		setArticleBody1, _  = jp.PackEvent("set_article_body", DummyEvSetArticleBody{"lorem ipsum ..."})
	)

	// Affixes
	var (
		affixOne, _ = jp.PackAffix(packing.Affix{"author/maxine": []packing.Hash{setAuthorName1.Hash()}})

		affixTwo, _ = jp.PackAffix(packing.Affix{"article/first": []packing.Hash{setArticleTitle1.Hash(), associateArticleAuthor1.Hash()}})

		affixThree, _ = jp.PackAffix(packing.Affix{"article/first": []packing.Hash{setArticleTitle2.Hash(), setArticleBody1.Hash()}})
	)

	// Checkpoints
	var (
		checkpointOne, _ = jp.PackCheckpoint(packing.Checkpoint{
			AffixHash:   affixOne.Hash(),
			CommandDesc: []byte(`{"create":"author"}`),
			Fields:      map[string]string{"session": "hello world"},
		})

		checkpointTwo, _ = jp.PackCheckpoint(packing.Checkpoint{
			AffixHash:    affixTwo.Hash(),
			CommandDesc:  []byte(`{"draft":"article"}`),
			Fields:       map[string]string{"session": "hello world"},
			ParentHashes: []packing.Hash{checkpointOne.Hash()},
		})

		checkpointThree, _ = jp.PackCheckpoint(packing.Checkpoint{
			AffixHash:    affixThree.Hash(),
			CommandDesc:  []byte(`{"update":"article"}`),
			Fields:       map[string]string{"session": "hello world"},
			ParentHashes: []packing.Hash{checkpointTwo.Hash()},
		})
	)

	tmpdir, err := ioutil.TempDir("", "depot_common_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	odbs := map[string]object.DB{
		"memory": &memory.ObjectStore{},
		"fs":     &fs.ObjectStore{BasePath: tmpdir},
	}
	refdbs := map[string]ref.DB{
		"memory": &memory.RefStore{},
		"fs":     &fs.RefStore{BasePath: tmpdir},
	}
	depots := map[string]types.Depot{
		"memory":    Simple{objdb: odbs["memory"], refdb: refdbs["memory"]},
		"fs":        Simple{objdb: odbs["fs"], refdb: refdbs["fs"]},
		"fs+memory": Simple{objdb: odbs["memory"], refdb: refdbs["fs"]},
		"memory+fs": Simple{objdb: odbs["fs"], refdb: refdbs["memory"]},
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
		refdb.Write("refs/heads/main", checkpointThree.Hash())
	}

	for name, depot := range depots {

		t.Run(name, func(t *testing.T) {
			t.Run("correctly checking for existence of aggregates", func(t *testing.T) {
				t.Skip("not implemented yet")
			})
		})

		t.Run(name, func(t *testing.T) {
			t.Run("correct events in correct order", func(t *testing.T) {

				var (
					wg            sync.WaitGroup
					ctx, cancelFn = context.WithTimeout(context.Background(), 1*time.Second)
				)
				wg.Add(5)

				go func() {
					wg.Wait()
					cancelFn()
				}()

				// Calling Glob() on a depot returns a PartitionIterator
				// a PartitionIterator's "Partitions" method returns a channel
				// of EventIterators and a cancellation function. The cancellation
				// function will self-cancel when the given Context expires.
				pIter := depot.Glob(ctx, "*")
				eIterCh, err := pIter.Partitions(ctx)
				if err != nil {
					panic(err) // TODO: can never happen, hard-coded nil
				}

				// An EventIterator comes with some metadata about the "partition" in question
				// and it's own way to emit events.
				for eIter := range eIterCh {
					go func(eIter types.EventIterator) {
						partitionEvs, err := eIter.Events(ctx)
						if err != nil {
							panic(err) // TODO: can never happen, hard-coded nil
						}
						for ev := range partitionEvs {
							fmt.Println(ev.Name(), ":", string(ev.Bytes()), "on", eIter.Pattern())
							wg.Done()
						}
					}(eIter)
				}
			})
		})
	}
}

// 	for name, depot := range depots {
// 		t.Run(name, func(t *testing.T) {
//
// 			t.Run("validation", func(t *testing.T) {
// 				t.Run("must refuse to store for paths not including an ID part, except _", func(t *testing.T) {
// 					t.Skip("not implemented yet")
// 				})
//
// 				t.Run("must allow events to survive a roundtrip of storage (incl args)", func(t *testing.T) {
// 					t.Skip("not implemented yet")
// 				})
// 			})
//
// 			t.Run("static-queries", func(t *testing.T) {
// 				t.Run("must allow lookup by verbatim path", func(t *testing.T) {
// 					t.Skip("not implemented yet")
// 				})
//
// 				t.Run("must allow lookup by globbing", func(t *testing.T) {
// 					t.Skip("not implemented yet")
// 				})
// 			})
//
// 			t.Run("rehydrate", func(t *testing.T) {
// 				t.Skip("must be able to rehydrate things")
// 			})
//
// 		})
// 	}
// }
