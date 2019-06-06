// +build unit

package queryable

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/ref"
	"github.com/retro-framework/go-retro/framework/matcher"
	"github.com/retro-framework/go-retro/framework/depot"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/retro"
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

type Predictable5sJumpClock struct {
	t     time.Time
	calls int
}

func (c *Predictable5sJumpClock) Now() time.Time {
	var next = c.t.Add(time.Duration((5 * c.calls)) * time.Second)
	c.calls = c.calls + 1
	return next
}

type alwaysMatches struct {}
func (_ alwaysMatches) DoesMatch(i interface{}) (bool, error) {
	fmt.Printf("checking for match on %s\n", i)
	return true, nil
}

func Test_Queryable(t *testing.T) {

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
				"session": "one",
				"date":    clock.Now().Format(time.RFC3339),
			},
		})

		checkpointTwo, _ = jp.PackCheckpoint(packing.Checkpoint{
			AffixHash:   affixTwo.Hash(),
			CommandDesc: []byte(`{"draft":"article"}`),
			Fields: map[string]string{
				"session": "one",
				"date":    clock.Now().Format(time.RFC3339),
			},
			ParentHashes: []retro.Hash{checkpointOne.Hash()},
		})

		checkpointThree, _ = jp.PackCheckpoint(packing.Checkpoint{
			AffixHash:   affixThree.Hash(),
			CommandDesc: []byte(`{"update":"article"}`),
			Fields: map[string]string{
				"session": "two",
				"date":    clock.Now().Format(time.RFC3339),
			},
			ParentHashes: []retro.Hash{checkpointTwo.Hash()},
		})

		// Extend
		checkpointFourA, _ = jp.PackCheckpoint(packing.Checkpoint{
			AffixHash:   affixFourA.Hash(),
			CommandDesc: []byte(`{"update":"article"}`),
			Fields: map[string]string{
				"session": "four",
				"date":    clock.Now().Format(time.RFC3339),
			},
			ParentHashes: []retro.Hash{checkpointThree.Hash()},
		})

		checkpointFourB, _ = jp.PackCheckpoint(packing.Checkpoint{
			AffixHash:   affixFourB.Hash(),
			CommandDesc: []byte(`{"update":"article"}`),
			Fields: map[string]string{
				"session": "four",
				"date":    clock.Now().Format(time.RFC3339),
			},
			ParentHashes: []retro.Hash{checkpointThree.Hash()},
		})

		_ = checkpointFourB
	)

	var populateDBs = func(odb object.DB, refdb ref.DB) (object.DB, ref.DB) {
		odb.WritePacked(setAuthorName1)
		odb.WritePacked(setArticleTitle1)
		odb.WritePacked(associateArticleAuthor1)
		odb.WritePacked(setArticleTitle2)
		odb.WritePacked(setArticleBody1)

		odb.WritePacked(affixOne)
		odb.WritePacked(affixTwo)
		odb.WritePacked(affixThree)
		odb.WritePacked(affixFourA)
		odb.WritePacked(affixFourB)

		odb.WritePacked(checkpointOne)
		odb.WritePacked(checkpointTwo)
		odb.WritePacked(checkpointThree)
		odb.WritePacked(checkpointFourA)
		odb.WritePacked(checkpointFourB)

		refdb.Write(depot.DefaultBranchName, checkpointFourA.Hash())
		refdb.Write(depot.DefaultBranchName+"alt", checkpointFourB.Hash())

		return odb, refdb
	}

	var queryableMatrix = map[string]func(object.DB, ref.DB) retro.Queryable{
		"depot": func(odb object.DB, refdb ref.DB) retro.Queryable {
			var o, r = populateDBs(odb, refdb)

			// Uncomment me to dump the hash table from the object db
			// if lsOdb, ok := o.(object.ListableSource); ok {
			// 	fmt.Println(lsOdb.Ls())
			// }

			return depot.NewSimple(o, r)
		},
	}

	var testCases = []struct {
		desc    string
		matcher retro.Matcher

		expectedResult []retro.URN
		expectedError  error
	}{
		{
			desc:           "can be found by the session ID on the default thread",
			matcher:        alwaysMatches{},
			expectedResult: []retro.URN{retro.URN("users/alice")},
		},
		// {
		// 	desc:           "can be found by the session ID on the default thread",
		// 	matcher:        matcher.NewSessionID("one"),
		// 	expectedResult: []retro.URN{retro.URN("users/alice")},
		// },
	}

	_ = matcher.NewSessionID("one")

	for querableName, queryableFn := range queryableMatrix {

		var queryable = queryableFn(&memory.ObjectStore{}, &memory.RefStore{})

		t.Run(querableName, func(t *testing.T) {
			for _, tc := range testCases {
				t.Run(tc.desc, func(t *testing.T) {
					res, err := queryable.Matching(context.TODO(), tc.matcher)
					fmt.Println(res, err)
				})
			}
		})
	}

}
