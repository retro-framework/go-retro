package widgets_app

import (
	"context"
	"errors"
	"io"

	"github.com/retro-framework/go-retro/aggregates"
	"github.com/retro-framework/go-retro/commands"
	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/retro"
)

// AllowCreationOfNewIdentities is used to toggle the creation of new
// identites on (effectively enabling signup) it may be redundant in the
// case of systems that use a SSO such as active directory or OAuth. An
// application instance that has never had this called may default to
// "false" subject to how it was initialized.
type AllowCreationOfNewIdentities struct {
	widgetsApp *aggregates.WidgetsApp
}

// SetState returns a WidgetsApp from the Aggregate that everyone else
// wants to deal with, every Aggregate type must implement this.
func (cmd *AllowCreationOfNewIdentities) SetState(agg retro.Aggregate) error {
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
func (cmd *AllowCreationOfNewIdentities) Apply(ctx context.Context, w io.Writer, session retro.Session, repo retro.Repo) (retro.CommandResult, error) {

	// identities := repo.Watch(ctx, "identities/*")
	// _ = identities
	// if !identities.HasAny() { // TODO: not implemented
	// 	return nil, errors.New("can't change application settings anonymously once identities exist")
	// }

	// if we are already allowing creation of identities
	// this command is a harmless noop
	if cmd.widgetsApp.AllowCreateIdentities == true {
		return nil, nil
	}

	// otherwise we allow creation of identities now, and emit
	// an event which attests to that fact in the future.
	// The aggregates.WidgetsApp must consider events.AllowCreationOfIdentities
	// in its ApplyTo() function.
	return retro.CommandResult{
		cmd.widgetsApp: []retro.Event{
			events.AllowCreateIdentities{},
		},
	}, nil
}

func init() {
	commands.Register(&aggregates.WidgetsApp{}, &AllowCreationOfNewIdentities{})
}
