package retro

import "context"

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
// Beware that failing to read some iterator implementations to the end
// may hold locks on some underlaying resources.
type EventIterator interface {
	Pattern() string
	Next(context.Context) (PersistedEvent, error)
	Events(context.Context) (<-chan PersistedEvent, <-chan error)
}
