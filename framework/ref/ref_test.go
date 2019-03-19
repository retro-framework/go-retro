// +build integration

package ref

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/retro"
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

	dbs := map[string]func() DB{
		"memory": func() DB { return &memory.RefStore{} },
		"fs": func() DB {
			os.RemoveAll(tmpdir)
			return &fs.RefStore{BasePath: tmpdir}
		},
	}

	var (
		fooHash = packing.HashStr("foo")
		barHash = packing.HashStr("bar")
	)

	for name, dbFn := range dbs {

		t.Run(name, func(t *testing.T) {

			t.Run("ensures that ref starts with refs/", func(t *testing.T) {
				t.Skip("not implemented yet")
			})

			t.Run("various read/write", func(t *testing.T) {

				var db = dbFn()

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

			t.Run("listing objects", func(t *testing.T) {

				var db = dbFn()

				_, err := db.Write("refs/heads/foo", fooHash)
				test.H(t).IsNil(err)
				_, err = db.Write("refs/heads/bar", barHash)
				test.H(t).IsNil(err)

				var want = map[string]retro.Hash{
					"refs/heads/bar": packing.HashStrToHash("sha256:fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9"),
					"refs/heads/foo": packing.HashStrToHash("sha256:2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae"),
				}

				if ldb, ok := db.(ListableStore); !ok {
					t.Skip(fmt.Sprintf("%s does not implement ListableStore interface", name))
				} else {
					res, err := ldb.Ls()
					test.H(t).IsNil(err)
					if diff := cmp.Diff(res, want); diff != "" {
						t.Errorf("results differs: (-got +want)\n%s", diff)
					}
				}

			})

		})

	}

}
