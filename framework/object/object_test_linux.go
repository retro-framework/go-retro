package object

import (
	"testing"

	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/storage/memory"
)

func Test_Store(t *testing.T) {

	stores := map[string]Store{
		"memory": &memory.ObjectStore{},
	}
	for name, store := range stores {
		t.Run(name, func(t *testing.T) {
			var (
				packedObj = packing.NewPackedObject("hello world")
			)
			t.Run("stores an object, returns the byte length on disk", func(t *testing.T) {
				len, err := store.WritePacked(packedObj)
				t.Log("len", len)
				t.Log("err", err)
			})
			t.Run("stores an object, returns zero length if already in store", func(t *testing.T) {
				len, err := store.WritePacked(packedObj)
				t.Log("len", len)
				t.Log("err", err)
			})

		})
	}

}
