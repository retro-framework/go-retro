package pack

import (
	"crypto/sha256"
	"fmt"
	"testing"

	test "github.com/retro-framework/go-retro/framework/test_helper"
)

type DummyEvent struct {
	Foo string `json:"foo"`
	Bar string `json:"bar"`
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
			wantContents = `event json dummy 29┋{"foo":"hello","bar":"world"}`
			wantHash     = `0def887a19c46e7d16ba7afc2e49f90b648207657dcccb4bd73444244180d9d4`
		)
		test.H(t).StringEql(string(res.Contents()), wantContents)
		test.H(t).StringEql(fmt.Sprintf("%x", res.Hash().Bytes), wantHash)
	})

	// TODO: check that affix tables are written in lexographical
	// key order
	// TODO: check what happens with exotic partition names (spaces,
	// special characters, etc)
	t.Run("exemplary affix", func(t *testing.T) {

		// Arrange
		jp := NewJSONPacker()
		hash := Hash{
			AlgoName: HashAlgoNameSHA256,
			Bytes:    sha256.New().Sum([]byte("foo")),
		}

		// Act
		res, err := jp.PackAffix(Affix{"baz/123": []Hash{hash}, "bar/123": []Hash{hash}})

		// Assert
		test.H(t).IsNil(err)
		var (
			wantContents = `affix 176┋0 bar/123 sha256:666f6fe3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
1 baz/123 sha256:666f6fe3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
`
			wantHash = `5f56d8826281ad53f09364366409ef9bf1118cb1da536c62a65e55b4ddef9af1`
		)
		test.H(t).StringEql(string(res.Contents()), wantContents)
		t.Log(fmt.Sprintf("%x", res.Hash().Bytes))
		test.H(t).StringEql(fmt.Sprintf("%x", res.Hash().Bytes), wantHash)
	})

}
