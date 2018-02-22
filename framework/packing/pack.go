package packing

import "errors"

const (
	HeaderContentSepRune = "\u0000"
)

var (
	ErrCheckpointWithoutAffix = errors.New("can not pack checkpoint without affix")
)

// Object is a storable object, they are
// created by
type Object interface {
	TypeName() ObjectTypeName
	Contents() []byte
}

// HashedObject is an object from the hash.
type HashedObject interface {
	Object
	Hash() Hash
}

// Event ns anything serializable for the future
type Event interface{}

// PartitionName as alias for string to make
// the documentation and code examples more
// speaking. Detail should not leak beyond the
// plumbing package.
type PartitionName string

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
// may prefer to ignore these events, but they do form par
// of our conceptual model.
type Affix map[PartitionName][]Hash

// A checkpoint represents a DDD command object execution
// and persistence of the resulting events. It stores
// an error incase the command failed.
type Checkpoint struct {
	Affix       HashedObject
	CommandDesc []byte
	Error       error
	Parents     []Checkpoint
	Fields      map[string]string
	Summary     string
	SessionID   string
}
