package types

import (
	"context"
	"time"
)

type NowFn func() time.Time

type Hash interface {
	String() string
}

type Aggregate interface {
	ReactTo(Event) error
}

// Event interface may be any type which may carry any baggage it likes.
// It must serialize and deserialize cleanly for storage reasons.
type Event interface{}

// EventName
type EventNameTuple struct {
	Name  string
	Event Event
}

type PersistedEvent interface {
	Time() time.Time
	Name() string
	Bytes() []byte
	CheckpointHash() Hash
	Event() (Event, error)
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

type CommandDesc interface {
	Name() string
	Path() string
}

type StateEngine interface {
	Apply(context.Context, SessionID, CommandDesc) (string, error)
	StartSession(SessionParams)
}

// Session is a type alias for Aggregate to make the function signatures
// more self-explanatory, feasibly in the future a session will be a
// superset of an Aggregate
type Session Aggregate

type SessionID string

type SessionParams interface {
	ID() SessionID
	Client() string
}

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
	Apply(context.Context, Aggregate, Depot) ([]Event, error)
	SetState(Aggregate) error
}

type CommandWithArgs interface {
	Command
	SetArgs(CommandArgs) error
}

type CommandArgs map[string]interface{}

type CommandFunc func(context.Context, Aggregate, Depot) ([]Event, error)

type Resolver interface {
	Resolve(context.Context, Depot, []byte) (CommandFunc, error)
}

type ResolveFunc func(context.Context, Depot, []byte) (CommandFunc, error)

type IDFactory func() (string, error)

type ObjectTypeName string

type HashedObject interface {
	Type() ObjectTypeName
	Contents() []byte
	Hash() Hash
}
