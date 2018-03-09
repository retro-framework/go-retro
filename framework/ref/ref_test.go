package ref

import (
	"crypto/sha256"
	"io/ioutil"
	"os"
	"testing"

	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/storage/fs"
	"github.com/retro-framework/go-retro/framework/storage/memory"
	test "github.com/retro-framework/go-retro/framework/test_helper"
)

func Test_DB(t *testing.T) {

	tmpdir, err := ioutil.TempDir("", "retro_framework_ref_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	dbs := map[string]DB{
		"memory": &memory.RefStore{},
		"fs":     &fs.RefStore{tmpdir},
	}

	for name, db := range dbs {
		t.Run(name, func(t *testing.T) {
			var (
				fooHash = packing.Hash{
					AlgoName: packing.HashAlgoNameSHA256,
					Bytes:    sha256.New().Sum([]byte("foo")),
				}
				barHash = packing.Hash{
					AlgoName: packing.HashAlgoNameSHA256,
					Bytes:    sha256.New().Sum([]byte("bar")),
				}
			)
			t.Run("ensures that ref starts with refs/", func(t *testing.T) {
				t.Skip()
			})
			t.Run("returns true when writing a ref for the first time", func(t *testing.T) {
				changed, err := db.Write("refs/heads/main", fooHash)
				test.H(t).IsNil(err)
				test.H(t).BoolEql(changed, true)
			})
			t.Run("returns false if hash is unchanged in store", func(t *testing.T) {
				changed, err := db.Write("refs/heads/main", fooHash)
				test.H(t).IsNil(err)
				test.H(t).BoolEql(changed, false)
			})
			t.Run("returns true if hash is changed overwritten in store", func(t *testing.T) {
				changed, err := db.Write("refs/heads/main", barHash)
				test.H(t).IsNil(err)
				test.H(t).BoolEql(changed, true)
			})
			t.Run("retrieves an existing object if already in store", func(t *testing.T) {
				packedHash, err := db.Retrieve("refs/heads/main")
				test.H(t).IsNil(err)
				test.H(t).StringEql(packedHash.String(), barHash.String())
			})
			t.Run("returns unknown ref error for non existent objects in store", func(t *testing.T) {
				_, err := db.Retrieve("refs/heads/non existent")
				test.H(t).NotNil(err)
			})
			t.Run("write symbolic fails if the ref name is not all caps", func(t *testing.T) {
				t.Skip("not implemented yet")
				_, err := db.WriteSymbolic("head", "refs/heads/mainline")
				test.H(t).NotNil(err)
			})
			t.Run("write symbolic returns true if the symbolic ref was created or changed", func(t *testing.T) {
				changed, err := db.WriteSymbolic("HEAD", "refs/heads/mainline")
				test.H(t).IsNil(err)
				test.H(t).BoolEql(changed, true)
			})
			t.Run("write symbolic returns false if the symbolic ref was created or changed", func(t *testing.T) {
				changed, err := db.WriteSymbolic("HEAD", "refs/heads/mainline")
				test.H(t).IsNil(err)
				test.H(t).BoolEql(changed, false)
			})
			t.Run("retrive symbolic returns symbolic ref correctly", func(t *testing.T) {
				ref, err := db.RetrieveSymbolic("HEAD")
				test.H(t).IsNil(err)
				test.H(t).StringEql(ref, "refs/heads/mainline")
			})
			t.Run("retrive symbolic returns error on non existent ref", func(t *testing.T) {
				_, err := db.RetrieveSymbolic("ANYTHING ELSE")
				test.H(t).NotNil(err)
			})
		})
	}

}