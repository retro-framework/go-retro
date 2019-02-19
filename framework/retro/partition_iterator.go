package retro

import "context"

// PartitionIterator iterates over matched Partitions in a consistent way
// it provides both a nested channel mechanism (Partitions) or a cursor
// approach ideal for a for p.Next() use-case.
type PartitionIterator interface {
	Pattern() string
	Next(context.Context) (EventIterator, error)
	Partitions(context.Context) (<-chan EventIterator, <-chan error)
}
