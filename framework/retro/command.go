package retro

import (
	"context"
	"io"
)

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
	Apply(context.Context, io.Writer, Session, Repo) (CommandResult, error)
	SetState(Aggregate) error
}
