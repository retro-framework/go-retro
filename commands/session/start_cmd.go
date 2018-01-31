package session

import (
	"context"
	"errors"

	"github.com/retro-framework/go-retro/aggregates"
	"github.com/retro-framework/go-retro/commands"
	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/types"
)

type Start struct {
	session *aggregates.Session
}

// State returns a WidgetsApp from the Aggregate that everyone else wants to
// deal with, every Aggregate type must implement this.
func (cmd *Start) SetState(agg types.Aggregate) error {
	if s, ok := agg.(*aggregates.Session); ok {
		cmd.session = s
		return nil
	} else {
		return errors.New("can't cast")
	}
}

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
func (cmd *Start) Apply(ctxt context.Context, _ types.Aggregate, aggStore types.Depot) ([]types.Event, error) {
	return []types.Event{events.StartSession{}}, nil
}

func init() {
	commands.Register(&aggregates.Session{}, &Start{})
}
