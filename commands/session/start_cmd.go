package session

import (
	"context"
	"errors"
	"io"

	"github.com/retro-framework/go-retro/aggregates"
	"github.com/retro-framework/go-retro/commands"
	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/retro"
)

// Start is used to gatekeep the creation of new sessions if starting a session
// (even anonymously) may have business value that should be reflected here,
// else a session aggregate if the application should otherwise be blocked, or
// only allow existing sessions this command could be modified to reject
// creation of new sessions, or create invalid sessions.
//
// Perhaps somewhat confusingly in case of session commands the command state
// receiver and the second (sesh) argument both refer to the same object. It is
// idiomatic to ignore the 2nd argument when operating on command sessions.
//
// Session->Start must put at least one event in the depot else subsequent commands
// to other aggregates will fail to find a session and raise an error.
type Start struct {
	session *aggregates.Session
}

// SetState receieves an anonymous Aggregate and must type assert
// it to the correct type (Session).
func (cmd *Start) SetState(agg retro.Aggregate) error {
	if s, ok := agg.(*aggregates.Session); ok {
		cmd.session = s
		return nil
	} else {
		return errors.New("can't cast")
	}
}

// Apply for sessions is effectively a noop in the default implementation
// it need only make a record in the data store that a session has been
// created and that we can look it up in the future.
func (cmd *Start) Apply(ctxt context.Context, _ io.Writer, _ retro.Session, repo retro.Repository) (retro.CommandResult, error) {
	return retro.CommandResult{
		cmd.session: []retro.Event{
			events.StartSession{},
		},
	}, nil
}

func init() {
	commands.Register(&aggregates.Session{}, &Start{})
}
