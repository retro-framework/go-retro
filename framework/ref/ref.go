package ref

import "github.com/retro-framework/go-retro/framework/packing"

// Store writes a named reference to a packing.Hash. It should return boolean
// whether the ref is now changed, and an error in case of storage problems.
type Store interface {
	Write(string, packing.Hash) (bool, error)
}

// Retrieve returns a packing hash given a symbolic name will return a
// packing.Hash pointer or an error.
type Source interface {
	Retrieve(string) (*packing.Hash, error)
}

type DB interface {
	Store
	Source
}
