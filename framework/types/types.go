package types

import "context"

type Aggregate interface {
	ReactTo(Event) error
}

type AggregateItterator interface {
	Len() (int, bool)
	Next() Aggregate
}

type Event interface{}

type EventItterator interface {
	Len() (int, bool)
	Next() Event
}

type Depot interface {
	Rehydrate(context.Context, Aggregate, string) error
	GetByDirname(context.Context, string) AggregateItterator
}

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
	Args() ApplicationCmdArgs
}
type ApplicationCmdArgs interface{}

type StateEngine interface {
	Apply(context.Context, SessionID, CommandDesc) (string, error)
	StartSession(SessionParams)
}

type SessionID string

type SessionParams interface{}

type AggregateManifest interface {
	Register(string, Aggregate) error
	ForPath(string) (Aggregate, error)
}

type EventManifest interface {
	Register(Event) error
}

type CommandManifest interface {
	Register(Aggregate, Command) error
	ForAggregate(Aggregate) ([]Command, error)
}

type Command interface {
	Apply(context.Context, Aggregate, Depot) ([]Event, error)
}

type CommandFunc func(context.Context, Aggregate, Depot) ([]Event, error)

type Resolver interface {
	Resolve(context.Context, Depot, []byte) (CommandFunc, error)
}

type ResolveFunc func(context.Context, Depot, []byte) (CommandFunc, error)
