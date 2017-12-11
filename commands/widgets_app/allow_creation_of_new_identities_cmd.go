package widgets_app

import (
	"context"
	"errors"

	"github.com/retro-framework/go-retro/aggregates"
	"github.com/retro-framework/go-retro/commands"
	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/types"
)

type AllowCreationOfNewIdentities struct {
	state types.Aggregate
}

// State returns a WidgetsApp from the Aggregate that everyone else
// wants to deal with, every Aggregate type must implement this.
func (cmd *AllowCreationOfNewIdentities) State() (*aggregates.WidgetsApp, error) {
	if wa, ok := cmd.state.(*aggregates.WidgetsApp); ok {
		return wa, nil
	} else {
		return &aggregates.WidgetsApp{}, errors.New("can't cast")
	}
}

// AllowCreationOfNewIdentities is used to toggle the creation of new
// identites on (effectively enabling signup) it may be redundant in the
// case of systems that use a SSO such as active directory or OAuth. An
// application instance that has never had this called may default to
// "false" subject to how it was initialized.
func (cmd *AllowCreationOfNewIdentities) Apply(ctxt context.Context, sesh types.Aggregate, aggStore types.Depot) ([]types.Event, error) {
	state, _ := cmd.State()

	// TODO: fix this to be sane, again
	// numIds, countable := repo.GetByDirname("identities").Len()
	// if !countable
	// 	return nil, errors.New("can't change application settings anonymously once identities exist")
	// }

	if state.AllowCreateIdentities == true {
		return nil, nil
	}

	return []types.Event{events.AllowCreateIdentities{}}, nil
}

func init() {
	commands.Register(&aggregates.WidgetsApp{}, &AllowCreationOfNewIdentities{})
}
