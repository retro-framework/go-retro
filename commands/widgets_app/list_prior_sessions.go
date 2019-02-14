package widgets_app

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/retro-framework/go-retro/aggregates"
	"github.com/retro-framework/go-retro/commands"
	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/types"
)

// AllowCreationOfNewIdentities is used to toggle the creation of new
// identites on (effectively enabling signup) it may be redundant in the
// case of systems that use a SSO such as active directory or OAuth. An
// application instance that has never had this called may default to
// "false" subject to how it was initialized.
type ListAllPriorSessions struct {
	widgetsApp *aggregates.WidgetsApp
}

// SetState returns a WidgetsApp from the Aggregate that everyone else
// wants to deal with, every Aggregate type must implement this.
func (cmd *ListAllPriorSessions) SetState(agg types.Aggregate) error {
	if wa, ok := agg.(*aggregates.WidgetsApp); ok {
		cmd.widgetsApp = wa
		return nil
	} else {
		return errors.New("can't cast")
	}
}

// Apply will check if there are any pre-existing identities in the depot and if
// there have not yet been any created, it will permit the creation of a new one.
//
// This allows configuration of the app early in its lifecycle.
func (cmd *ListAllPriorSessions) Apply(ctx context.Context, w io.Writer, session types.Session, depot types.Depot) (types.CommandResult, error) {

	fmt.Fprint(w, "Hello World!")

	return types.CommandResult{
		cmd.widgetsApp: []types.Event{
			events.PriorSessionsListed{},
		},
	}, nil
}

func init() {
	commands.Register(&aggregates.WidgetsApp{}, &ListAllPriorSessions{})
}
