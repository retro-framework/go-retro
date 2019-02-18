package types

import (
	"context"
	"io"
	"time"
)

// RefMove represents a head pointer movement
// it contains the old and new hashes. A boolean
// is set indicating whether this is a FF move
// or not.
type RefMove struct {
	Old Hash
	New Hash
	FF  bool
}

// Hash is a minimal interface for the framework
// as a whole. It can broadly be considered as
// a type alias for string as most of the storage
// and "downstream" (to clients) code works with
// serialized data. There is a concrete implementation
// of hash which offers more methods but String()
// is lingua franca.
type Hash interface {
	String() string
	Bytes() []byte
}

// Aggregate interface is simple, it need only ReactTo
// events. This Aggregate interface is only relevant to the
// write side of the application.
//
// Name is used by the Command objects to get a name for
// aggregates. If aggregates are anonymous then they may
// also receive SetName. Renaming aggregates is not supported
type Aggregate interface {
	ReactTo(Event) error

	// TODO: pull these out into a separate interface
	// as below so that we can keep this reserved
	// for the Engine internals.
	Name() PartitionName
	SetName(PartitionName) error
}

// NamedAggregate is used by the Engine to stealthily pass
// names through to the command, where appropriate. The
// interface should never be used in a Command.
//
// For known, loaded-from-storage aggregates the name can
// be set in the engine and survive the roundtrip to the
// command and be read back in the Engine after Apply has
// returned. The CommandResult map[types.Aggregage][]types.Event
// then allows the use of the interface via upgrade on
// the keys of the map to get the names to use when persisting
// the events.
type ExistingAggregate interface {
	Name() PartitionName
}

// Event interface may be any type which may carry any baggage it likes.
// It must serialize and deserialize cleanly for storage reasons.
type Event interface{}

// EventNameTuple is used exclusively in the tests for constructing
// a test fixture. It could be moved into conditional compilation or
// defined solely in the tests potentially.
type EventNameTuple struct {
	Name  string
	Event Event
}

// Clock allows dependency injection of a function returning
// the current time. Due to all the test code dealing with serialized
// data a clock with a predictable step is used extensively.
type Clock interface {
	Now() time.Time
}

// PersistedEvent mirrors an event embellishing it with the data
// inferred from its place within the Merkle Tree. The raw data
// returned by Bytes() will need to be handled with an EventManifest
// and a name lookup table to instantiate it back to a real concrete type
type PersistedEvent interface {
	Time() time.Time
	Name() string
	Bytes() []byte
	CheckpointHash() Hash
	PartitionName() PartitionName
}

// Logger is the generic logging interface. It explicitly avoids including
// Fatal and Fatalf because of the relative brutal nature of os.Exit
// without a chance to clean up.
//
// In general tracing should be preferred to logging, however logging can
// always be valuable.
type Logger interface {
	Debug(...interface{})
	Debugf(string, ...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Warn(...interface{})
	Warnf(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
}

// CommandDesc is used after parsing the raw byte stream from downstream
// clients. A CommandDesc is used internally as soon as the JSON parsing
// is done by the superficial request handling layer.
type CommandDesc interface {
	Name() string
	Path() string
}

// Session is a type alias for Aggregate to make the function signatures
// more self-explanatory, feasibly in the future a session will be a
// superset of an Aggregate
type Session Aggregate

// SessionID is a type alias for string to convey the meaning that a real
// session ID is required and not any (maybe empty) string. In the future
// the interface may be broadened to make it behave more like a real
// type with methods to access commonly required information.
type SessionID string

// Command is the generic interface to express a user intent towards a
// model in the system.
//
// Commands exist to carry state, the primary calling method is to pass a
// reference to the Apply() function to the calling site, our public
// interface then simply demands that we can pass a simple function, not
// the entire object (closures ensure that the object context is always
// available)
//
// SetState is used to infuse the command with Aggregate state to whom it
// is attached. The aggregate state is embedded into the struct
// implementing Command rathe than given as an argument to express that
// logically one is calling a method *on* an aggregate that has been
// brought upto a certain state.
//
// Apply takes a context, an Aggregate which is expected to represent the
// current session, and a Depot which it may use to look up any other
// Aggregates that it needs to apply business logic.
type Command interface {
	Apply(context.Context, io.Writer, Session, Repository) (CommandResult, error)
	SetState(Aggregate) error
}

// CommandWithArgs is an optional interface upgrade on Command
// which exposes a new SetArgs command which can be used to pass
// user data (parsed out of {params: ...} in the CommandDesc)
type CommandWithArgs interface {
	Command
	SetArgs(CommandArgs) error
}

type CommandWithRenderFn interface {
	Render(context.Context, io.Writer, Session, CommandResult) error
}

// CommandArgs is a type alias for interface{}
// to express when we are dealing with CommandArgs
// and not a real anon interface.
type CommandArgs interface{}

// CommandFunc is the main heavy-lifting of a Command. The CommandFunc
// is easier to use in tests where there may be no need for heavy
// boilerplate code.
type CommandFunc func(context.Context, io.Writer, Session, Repository) (CommandResult, error)

// CommandResult is a type alias for map[string][]Event
// to make the function signatures expressive. The resulting
// map should ideally contain
type CommandResult map[Aggregate][]Event

// EventFixture is a type alias for testing, it aliases CommandResult
// as the mapping of Aggregate to []Event is useful for testing
type EventFixture CommandResult

// Resolver takes a []byte and returns a callable command function
// the resolver is used bt the Engine to take serialized client
// input and map it to a registered command by name. The command
// func returned will usually be a function on a struct type
// which the resolver will instantiate and prepare for execution.
type Resolver interface {
	Resolve(context.Context, Repository, []byte) (Command, error)
}

// ResolveFunc does the heavy lifting on the resolution. The Resolver
// interface is clumbsy for use in tests and the ResolveFunc allows
// a simple anonymous drop-in in tests which can resolve a stub/double
// without lots of boilerplate code.
type ResolveFunc func(context.Context, Repository, []byte) (Command, error)

// IDFactory is a function that should generate IDs. This is primarily
// used in the Engine implementations to generate an ID for newly created
// things where no ID has been provided.
type IDFactory func() (string, error)

// ObjectTypeName is a pseudo enum for the known types of blob stored
// in the object database.
type ObjectTypeName string

// HashedObject is a simple interface allowing more than one type
// of object to be hashed without knowing the type up-front.
// consumers can switch on the result of Type() and cast explicitly
// to one or the other type.
type HashedObject interface {
	Type() ObjectTypeName
	Contents() []byte
	Hash() Hash
}

// PatternMatcher defines a single function interface
// for matching patterns. It is used to compare the aggregate
// paths within an affix to the aggregate name being searched
// for. In a sane implementation it should support at least
// POSIX globbing and perhaps even Regular Expressions to
// allow for matching such as `users/*` or similar.
//
// In testing, this pattern matcher may be replaced with a
// no-op or static matcher.
type PatternMatcher interface {
	DoesMatch(pattern, partition string) (bool, error)
}
