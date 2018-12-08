package object

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/storage/fs"
	"github.com/retro-framework/go-retro/framework/storage/memory"
	test "github.com/retro-framework/go-retro/framework/test_helper"
)

// TODO: I think the obj store should do the hashing itself, you just hand over
// something packed (which may be a type alias for a []byte)
func Test_DB(t *testing.T) {

	tmpdir, err := ioutil.TempDir("", "retro_framework_object_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	dbs := map[string]DB{
		"memory": &memory.ObjectStore{},
		"fs":     &fs.ObjectStore{BasePath: tmpdir},
	}

	for name, db := range dbs {
		t.Run(name, func(t *testing.T) {
			var (
				packedObj = packing.NewPackedObject("hello world")
			)
			t.Run("stores an object returns the byte length at rest", func(t *testing.T) {
				len, err := db.WritePacked(packedObj)
				test.H(t).IsNil(err)
				test.H(t).IntEql(len, 23)
			})
			t.Run("stores an object returns zero length if already in store", func(t *testing.T) {
				len, err := db.WritePacked(packedObj)
				test.H(t).IsNil(err)
				test.H(t).IntEql(len, 0)
			})
			t.Run("retrieves an existing object if already in store", func(t *testing.T) {
				po, err := db.RetrievePacked(packedObj.Hash().String())
				test.H(t).IsNil(err)
				test.H(t).StringEql(string(packedObj.Contents()), string(po.Contents()))
			})
			t.Run("errors when retriving a object not already in the store", func(t *testing.T) {
				t.Skip("not implemented yet")
			})
		})
	}

}
