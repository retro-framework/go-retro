package packing

import (
	"github.com/retro-framework/go-retro/framework/types"
)

const (
	// HeaderContentSepRune is set to the null byte. This is used
	// as a field separator in the binary packed representations
	// of packed objects. The separator could be anything, the
	// null byte is used as a historical nod to binary protocols
	// and because it's virtually the only valid byte that can
	// never be part of a meaningful input.
	HeaderContentSepRune = "\u0000"
)

// Object is a storable object, they are
// created by
type Object interface {
	Contents() []byte
}

// Event ns anything serializable for the future
type Event interface{}

// Affix is a map of partition names to slices of events.
// Affixes are closely related to checkpoints. If a command
// emits a bunch of related events they will be packed into
// a single affix and it will be clear that they were emitted
// at the same time.
//
// An affix *may* be completely empty, or a partition's event
// list may be empty. A failed command execution may yield
// some events, but also an error in which case we would get
// a partial affix, but checkpoint it with an error. The reader
// may prefer to ignore these events, but they do form part
// of our conceptual model.
type Affix map[types.PartitionName][]types.Hash
