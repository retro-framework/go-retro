package depot

import (
	"testing"

	"github.com/retro-framework/go-retro/framework/packing"
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

	jp := packing.NewJSONPacker()

	// Events
	var (
		set_author_name_1, _ = jp.PackEvent("set_author_name", DummyEvSetAuthorName{"Maxine Mustermann"})

		set_article_title_1, _        = jp.PackEvent("set_article_title", DummyEvSetArticleTitle{"event graph for noobs"})
		associate_article_author_1, _ = jp.PackEvent("associate_article_author", DummyEvAssociateArticleAuthor{"author/maxine"})

		set_article_title_2, _ = jp.PackEvent("set_article_title", DummyEvSetArticleTitle{"learning event graph"})
		set_article_body_1, _  = jp.PackEvent("set_article_body", DummyEvSetArticleBody{"lorem ipsum ..."})
	)

	// Affixes
	var (
		affix_1, _ = jp.PackAffix(packing.Affix{"author/maxine": []packing.Hash{set_author_name_1.Hash()}})

		affix_2, _ = jp.PackAffix(packing.Affix{"article/first": []packing.Hash{set_article_title_1.Hash(), associate_article_author_1.Hash()}})

		affix_3, _ = jp.PackAffix(packing.Affix{"article/first": []packing.Hash{set_article_title_2.Hash(), set_article_body_1.Hash()}})
	)

	// Checkpoints
	var (
		first_checkpoint, _ = jp.PackCheckpoint(packing.Checkpoint{
			AffixHash:   affix_1.Hash(),
			CommandDesc: []byte(`{"create":"author"}`),
			Fields:      map[string]string{"session": "hello world"},
		})

		second_checkpoint, _ = jp.PackCheckpoint(packing.Checkpoint{
			AffixHash:    affix_2.Hash(),
			CommandDesc:  []byte(`{"draft:"article"}`),
			Fields:       map[string]string{"session": "hello world"},
			ParentHashes: []packing.Hash{first_checkpoint.Hash()},
		})

		third_checkpoint, _ = jp.PackCheckpoint(packing.Checkpoint{
			AffixHash:    affix_3.Hash(),
			CommandDesc:  []byte(`{"update":"article"}`),
			Fields:       map[string]string{"session": "hello world"},
			ParentHashes: []packing.Hash{second_checkpoint.Hash()},
		})
	)

	_ = third_checkpoint

	// depots := map[string]types.Depot{
	// 	"in-memory": memory.NewEmptyDepot(),
	// 	"redis":     redis.NewDepot(),
	// }
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
