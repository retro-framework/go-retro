package packing

import (
	"crypto/sha256"
	"hash"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	test "github.com/retro-framework/go-retro/framework/test_helper"
)

type DummyEvent struct {
	Foo string `json:"foo"`
	Bar string `json:"bar"`
}

func Test_UnpackPack(t *testing.T) {

	t.Run("exemplary event", func(t *testing.T) {

		// Arrange
		jp := NewJSONPacker()

		// Act
		packed, _ := jp.PackEvent("dummy", DummyEvent{"hello", "world"})
		name, payload, err := jp.UnpackEvent(packed.Contents())

		// Assert
		test.H(t).IsNil(err)
		test.H(t).StringEql(name, "dummy")
		test.H(t).StringEql(string(payload), `{"foo":"hello","bar":"world"}`)

	})

	t.Run("exemplary affix", func(t *testing.T) {

		// Arrange
		var (
			jp   = NewJSONPacker()
			hash = hashStr("foo")
		)

		// Act
		packed, _ := jp.PackAffix(Affix{"baz/123": []Hash{hash}, "bar/123": []Hash{hash}})
		aggregateEvHashes, err := jp.UnpackAffix(packed.Contents())

		// Assert
		test.H(t).IsNil(err)
		var want = map[PartitionName][]string{
			PartitionName("bar/123"): []string{"sha256:2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae"},
			PartitionName("baz/123"): []string{"sha256:2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae"},
		}
		if cmp.Equal(aggregateEvHashes, want) != true {
			t.Fatalf("equality assertion failed: %s", cmp.Diff(aggregateEvHashes, want))
		}
	})

	t.Run("exemplary checkpoint", func(t *testing.T) {
		t.Skip("not implemented yet")
	})
}

func Test_Pack(t *testing.T) {

	t.Run("exemplary event", func(t *testing.T) {

		// Arrange
		jp := NewJSONPacker()

		// Act
		res, err := jp.PackEvent("dummy", DummyEvent{"hello", "world"})

		// Assert
		test.H(t).IsNil(err)
		var (
			wantContents = `event json dummy 29` + HeaderContentSepRune + `{"foo":"hello","bar":"world"}`
			wantHash     = `sha256:0756fae7f4a43d60b5532e1d4da5665daeb0f1a5274f363b99a7757511ec88db`
		)
		test.H(t).StringEql(string(res.Contents()), wantContents)
		test.H(t).StringEql(res.Hash().String(), wantHash)
	})

	// TODO: check that affix tables are written in lexographical
	// key order
	// TODO: check what happens with exotic partition names (spaces,
	// special characters, etc)
	t.Run("exemplary affix", func(t *testing.T) {

		// Arrange
		var (
			jp   = NewJSONPacker()
			hash = hashStr("foo")
		)

		// Act
		res, err := jp.PackAffix(Affix{"baz/123": []Hash{hash}, "bar/123": []Hash{hash}})

		// Assert
		test.H(t).IsNil(err)
		var (
			wantContents = `affix 164` + HeaderContentSepRune + `0 bar/123 sha256:2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae
1 baz/123 sha256:2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae
`
			wantHash = `sha256:f51b6e929c8f79300603460b9d545c51993cfc5c7d2f05808357baec29f84a4d`
		)
		test.H(t).StringEql(string(res.Contents()), wantContents)
		test.H(t).StringEql(res.Hash().String(), wantHash)
	})

	// TODO: With more than one parent
	// TODO: With fields (?)
	// TODO: With summary
	// TODO: With no session ID
	// TODO: With more than one (sorted) field(s)?
	// TODO: Error (in body)
	t.Run("exemplary parentless checkpoint", func(t *testing.T) {

		// Arrange
		var (
			jp = &JSONPacker{
				hashFn: func() hash.Hash { return sha256.New() },
				nowFn:  func() time.Time { return time.Time{} },
			}
			hash = hashStr("hello")
		)

		// Act
		checkpoint := Checkpoint{
			AffixHash:    hash,
			CommandDesc:  []byte(`{"foo":"bar"}`),
			Error:        nil,
			Fields:       map[string]string{"session": "hello world"},
			ParentHashes: []Hash{},
			SessionID:    "hello world",
		}
		res, err := jp.PackCheckpoint(checkpoint)

		// Assert
		test.H(t).IsNil(err)
		var (
			wantContents = `checkpoint 113` + HeaderContentSepRune + `affix sha256:2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae
session hello world

{"foo":"bar"}
`
			wantHash = `sha256:922d43b9cfa0911fdb69dc17aab1221e9906c2343a0db876b18c555c1aef8da0`
		)
		test.H(t).StringEql(string(res.Contents()), wantContents)
		test.H(t).StringEql(res.Hash().String(), wantHash)
	})

}
