package ref

import (
	"github.com/retro-framework/go-retro/framework/retro"
)

// Store writes a named reference to a packing.Hash. It should return boolean
// whether the ref is now changed, and an error in case of storage problems.
type Store interface {
	Write(string, retro.Hash) (bool, error)
	WriteSymbolic(string, string) (bool, error)
}

// ListableStore is optionally implementable by objects otherwise conforming to
// the Store interface. Not all stores are listable, and the implementation and
// signature of the optional ListableStore.Ls() make no guarantees or recommendations
// on the handling of collation or duplicates, etc.
type ListableStore interface {
	Store
	Ls() (map[string]retro.Hash, error)
}

// Source interface returns a packing hash given a symbolic name will return a
// packing.Hash pointer or an error.
type Source interface {
	Retrieve(string) (retro.Hash, error)
	RetrieveSymbolic(string) (string, error)
}

// DB is a combination of a Source and a Store. Both must be implemented
// for a database to be usable. The interfaces are split for testing purposes
// and to reflect the potential asymetry in storage and retrieval inherant
// in any system. One may wish to implement a DB which uses a transparent
// read through cache and passes writes to durable storage - for this a split
// interface makes perfect sense.
type DB interface {
	Store
	Source
}
