package storage

import "github.com/leehambley/ls-cms/events"

// Depot comprises a EventStore and CommandStore, it is
// expected to conform to the Repository interface so that
// events originating from a command can be stored and
// associated with that command
type Depot Repository

// Hydrator is the flipped side of the coin for aggregates
// represented here as we're solely interested in storage
// concerns, and to avoid circular dependencies.
//
// Hydrated is any type that can satisfy yielding it's stored events
// itteratively. Should send io.EOF on error.
//
// Effectively this is a "hydrated" aggregate, it's the only interface we
// care about, as it's our job to dry it out.
type Hydrated interface {
	ReactTo(events.Event) error
	// NextEv(startOffset int) (events.Event, error)
}

type HydratedItterator interface {
	Len() (int, bool)
	Next() Hydrated
}

// EventStore is responsible for storing the aggregate state Aggregates are
// individual objects (actors, if you like) in the system, see:
//
// Path is something resembling the "path' in a url such as: users/123
// underscore is a valid path and is used for "unresolveable" events
// (targeted at master)
//
// When the path dirname is _ it is a special case for the "root" object,
// e.g the sole application instance. There may be a way to do multi
// tennancy here, but I don't know that it makes sense.
//
// GetByDirname gives us back all users given something like "users"
//
// https://lostechies.com/gabrielschenker/2015/06/06/event-sourcing-applied-the-aggregate/
type Repository interface {
	Rehydrate(Hydrated, string) error
	GetByDirname(string) HydratedItterator
}
