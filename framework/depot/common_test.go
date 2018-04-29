package depot

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	// "github.com/fortytw2/leaktest"

	"github.com/google/go-cmp/cmp"
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

	var evNameTuples = []types.EventNameTuple{
		{Name: "set_author_name", Event: DummyEvSetAuthorName{"Maxine Mustermann"}},
		{Name: "set_article_title", Event: DummyEvSetArticleTitle{"event graph for noobs"}},
		{Name: "associate_article_author", Event: DummyEvAssociateArticleAuthor{"author/maxine"}},
		{Name: "set_article_title", Event: DummyEvSetArticleTitle{"learning event graph"}},
		{Name: "set_article_body", Event: DummyEvSetArticleBody{"lorem ipsum ..."}},
	}

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
		affixOne, _ = jp.PackAffix(packing.Affix{"author/maxine": []types.Hash{setAuthorName1.Hash()}})

		affixTwo, _ = jp.PackAffix(packing.Affix{"article/first": []types.Hash{setArticleTitle1.Hash(), associateArticleAuthor1.Hash()}})

		affixThree, _ = jp.PackAffix(packing.Affix{"article/first": []types.Hash{setArticleTitle2.Hash(), setArticleBody1.Hash()}})
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
			ParentHashes: []types.Hash{checkpointOne.Hash()},
		})

		checkpointThree, _ = jp.PackCheckpoint(packing.Checkpoint{
			AffixHash:    affixThree.Hash(),
			CommandDesc:  []byte(`{"update":"article"}`),
			Fields:       map[string]string{"session": "hello world"},
			ParentHashes: []types.Hash{checkpointTwo.Hash()},
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
		"memory": Simple{objdb: odbs["memory"], refdb: refdbs["memory"]},
		// "fs":        Simple{objdb: odbs["fs"], refdb: refdbs["fs"]},
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
			t.Run("correct events in correct order", func(t *testing.T) {

				t.Skip("nothing to do here")

				var seenEventTuples []types.EventNameTuple
				var ctx, cancelFn = context.WithTimeout(context.Background(), 5*time.Second)
				defer cancelFn()

				matching := depot.Glob(ctx, "*")
				matchedPartitions, matcherErrors := matching.Partitions(ctx)
			Out:
				for {
					select {
					case <-ctx.Done():
						fmt.Println("test timed out")
						break Out
					case err := <-matcherErrors:
						if err != nil {
							t.Fatalf("matcher error %q", err)
						}
					case partition, stillOpen := <-matchedPartitions:
						if !stillOpen {
							fmt.Println("matched partitions is closed")
							matchedPartitions = nil
							break
						}
						fmt.Println("got an event iterator")
						events, partitionErrors := partition.Events(ctx)
						for {
							select {
							case err := <-partitionErrors:
								if err != nil {
									t.Fatalf("partition error %q", err)
								}
							case ev, stillOpen := <-events:
								fmt.Println("got an event", ev)
								if !stillOpen {
									fmt.Println("events ch is closed")
									events = nil
									break
								}
								fmt.Println("ReceivedEv", ev)
							}
						}
					}
				}
				if diff := cmp.Diff(evNameTuples, seenEventTuples); diff != "" {
					t.Errorf("results differs: (-got +want)\n%s", diff)
				}
			})
		})
	}
}
