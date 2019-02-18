package identity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/retro-framework/go-retro/aggregates"
	"github.com/retro-framework/go-retro/commands"
	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/types"
)

type HideIdentity struct {
	identity *aggregates.Identity
}

// SetState receieves an anonymous Aggregate and must type assert
// it to the correct type (Identity).
func (cmd *HideIdentity) SetState(agg types.Aggregate) error {
	if typedAggregate, ok := agg.(*aggregates.Identity); ok {
		cmd.identity = typedAggregate
		return nil
	} else {
		return errors.New("can't cast")
	}
}

func (cmd *HideIdentity) Apply(ctxt context.Context, w io.Writer, session types.Session, repo types.Repository) (types.CommandResult, error) {

	s := session.(*aggregates.Session)
	if !s.HasIdentity {
		return nil, fmt.Errorf("you haven't authenticated or you don't have an identity")
	}
	if s.IdentityName != cmd.identity.Name() {
		return nil, fmt.Errorf("Identity name on session does not match, this is not your profile")
	}

	json.NewEncoder(w).Encode(cmd.identity)

	return types.CommandResult{
		cmd.identity: []types.Event{
			events.SetVisibility{Radius: "private"},
		},
	}, nil
}

func init() {
	commands.Register(&aggregates.Identity{}, &HideIdentity{})
}