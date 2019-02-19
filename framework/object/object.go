package object

import (
	"github.com/retro-framework/go-retro/framework/retro"
)

// Store takes a packed object and stores it, it may apply some
// disk compression. A 0 byte return indicate that the object was
// not stored, but if no error is returned, that this object was
// already in storage. Packed objects are not compressed, they
// are simply packed bytes with an associated hash.
type Store interface {
	WritePacked(retro.HashedObject) (int, error)
}

// Source takes a string in the format of a hash with prefix (e.g
// sha256:b937....19251876f7) and returned a Hashed object which can
// be parsed by the caller.
type Source interface {
	RetrievePacked(string) (retro.HashedObject, error)
}

type ListableSource interface {
	Ls() []retro.Hash
}

type DB interface {
	Store
	Source
}
