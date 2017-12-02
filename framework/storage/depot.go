package storage

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
