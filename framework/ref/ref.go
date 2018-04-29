package ref

import (
	"errors"

	"github.com/retro-framework/go-retro/framework/types"
)

var (
	ErrUnknown         = errors.New("ref: ref unknown")
	ErrUnknownSymbolic = errors.New("ref: symbolic ref unknown")
)

// Store writes a named reference to a packing.Hash. It should return boolean
// whether the ref is now changed, and an error in case of storage problems.
type Store interface {
	Write(string, types.Hash) (bool, error)
	WriteSymbolic(string, string) (bool, error)
}

// Retrieve returns a packing hash given a symbolic name will return a
// packing.Hash pointer or an error.
type Source interface {
	Retrieve(string) (types.Hash, error)
	RetrieveSymbolic(string) (string, error)
}

type DB interface {
	Store
	Source
}
