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
			jp    = NewJSONPacker()
			hash  = hashStr("foo")
			affix = Affix{"baz/123": []Hash{hash}, "bar/123": []Hash{hash}}
		)

		// Act
		packed, _ := jp.PackAffix(affix)
		unpackedAffix, err := jp.UnpackAffix(packed.Contents())

		// Assert
		test.H(t).IsNil(err)
		if cmp.Equal(unpackedAffix, affix) != true {
			t.Fatalf("equality assertion failed: %s", cmp.Diff(unpackedAffix, affix))
		}
	})

	t.Run("exemplary checkpoint", func(t *testing.T) {

		var (
			jp = NewJSONPacker()

			affixHash      = hashStr("affix")
			checkpointHash = hashStr("checkpoint")
			checkpoint     = Checkpoint{
				AffixHash:    affixHash,
				CommandDesc:  []byte(`{"foo":"bar"}`),
				Fields:       map[string]string{"session": "DEADBEEF-SESSIONID"},
				ParentHashes: []Hash{checkpointHash},
			}
		)

		packed, _ := jp.PackCheckpoint(checkpoint)
		unpackedCheckpoint, err := jp.UnpackCheckpoint(packed.Contents())

		// Assert
		test.H(t).IsNil(err)
		if cmp.Equal(unpackedCheckpoint, checkpoint) != true {
			t.Fatalf("equality assertion failed: %s", cmp.Diff(unpackedCheckpoint, checkpoint))
		}

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
			Fields:       map[string]string{"session": "DEADBEEF-SESSIONID"},
			ParentHashes: []Hash{},
		}
		res, err := jp.PackCheckpoint(checkpoint)

		// Assert
		test.H(t).IsNil(err)
		var (
			wantContents = `checkpoint 120` + HeaderContentSepRune + `affix sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
session DEADBEEF-SESSIONID

{"foo":"bar"}
`
			wantHash = `sha256:7903903299257e21884b0fd0f5fa50145bdf47ef0b0ffac60dbf52642b758044`
		)
		test.H(t).StringEql(string(res.Contents()), wantContents)
		test.H(t).StringEql(res.Hash().String(), wantHash)
	})

}
