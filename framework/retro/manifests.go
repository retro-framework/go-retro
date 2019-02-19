package retro

type AggregateManifest interface {
	Register(string, Aggregate) error
	ForPath(string) (Aggregate, error)
}

type ListingAggregateManifest interface {
	List() []string
}

type EventManifest interface {
	Register(Event) error
	RegisterAs(string, Event) error
	KeyFor(Event) (string, error)
	ForName(string) (Event, error)
}

type ListingEventManifest interface {
	List() map[string]interface{}
}

type EventFactory interface {
	New(string) Event
}

// CommandManifest is the interface that allows for various implementations
// of mapping command types to aggregates. They are stored internally in a
// map of types to a slice of strings.
//
// Register takes an aggregate and derives it's type and appends the
// Command type to a slice. There is room for alternative implementations
// which may be faster, or do a smarter search than the range loop to find
// matching commands.
//
// ForAggregate is counterpart to Register, it returns the Commands ready
// to apply, or an error.
type CommandManifest interface {
	Register(Aggregate, Command) error
	ForAggregate(Aggregate) ([]Command, error)

	RegisterWithArgs(Aggregate, Command, interface{}) error
	ArgTypeFor(Command) (CommandArgs, bool)
}

type ListingCommandManifest interface {
	List() map[string][]string
}
