package types

import "context"

// EventCursor is a convenience for iterating over
// events within a partition without concerning
// the caller with channel semantics
type EventCursor interface {
	ParitionName() string
	Value() PersistedEvent

	Next() bool
	Err() error
	Close()
}

// PartitionCursor is a convenience for iterating over
// events within a partition without concerning
// the caller with channel semantics
type PartitionCursor interface {
	Value() EventCursor

	Next() bool
	Err() error
	Close()
}

// PartitionIterator iterates over matched Partitions in a consistent way
// it provides both a nested channel mechanism (Partitions) or a cursor
// approach ideal for a for p.Next() use-case.
type PartitionIterator interface {
	Pattern() string
	Partitions(context.Context) (<-chan EventIterator, <-chan error)
	// PartitionCursor(context.Context) PartitionCursor
}

// EventIterator is a simple Iterator interface which should mean that we
// are never in a situation where an aggregate with a large number of
// events causes massive allocations or other resource starvation when
// being rehydrated.
//
// Implementations may choose to batch their reads to the underlaying
// storage into bulk, and iterate over single items at the API level, this
// has been shown to be very performant when talking to Redis for example
// in "pages" of 1000.
//
// Beware that failing to read some itterator implementations to the end
// may hold locks on some underlaying resources.
type EventIterator interface {
	Pattern() string
	Events(context.Context) (<-chan PersistedEvent, <-chan error)
}
