package packing

// import (
// 	"crypto/sha256"
// 	"hash"
// 	"testing"
// 	"time"
//
// 	test "github.com/retro-framework/go-retro/framework/test_helper"
// )
//
// func Test_PackedObject(t *testing.T) {
//
// 	t.Run("with no parents", func(t *testing.T) {
//
// 		// Arrange
// 		var (
// 			jp = &JSONPacker{
// 				hashFn: func() hash.Hash { return sha256.New() },
// 				nowFn:  func() time.Time { return time.Time{} },
// 			}
// 			hash = Hash{
// 				AlgoName: HashAlgoNameSHA256,
// 				Bytes:    sha256.New().Sum([]byte("foo")),
// 			}
// 			packedAffix, _ = jp.PackAffix(Affix{"baz/123": []Hash{hash}, "bar/123": []Hash{hash}})
// 		)
//
// 		// Act
// 		checkpoint := Checkpoint{
// 			Affix: packedAffix,
// 			// Summary:     "test checkpoint",
// 			CommandDesc: []byte(`{"foo":"bar"}`),
// 			Error:       nil,
// 			Fields:      map[string]string{"session": "hello world"},
// 			Parents:     []Checkpoint{},
// 			SessionID:   "hello world",
// 		}
// 		res, err := jp.PackCheckpoint(checkpoint)
//
// 		// Assert
// 		test.H(t).IsNil(err)
// 		var (
// 			wantContents = `checkpoint 113` + HeaderContentSepRune + `affix sha256:b9371e220f8a4c8fe071a6e7d7b2e6788f243ba2f88553a15e258219251876f7
// session hello world
//
// {"foo":"bar"}
// `
// 			wantHash = `sha256:e830337714408778866b7111778c79c4d437ddd48008169dc4d7a44484f2aeee`
// 		)
// 		test.H(t).StringEql(string(res.Contents()), wantContents)
// 		test.H(t).StringEql(res.Hash().String(), wantHash)
// 	})
//
// }
