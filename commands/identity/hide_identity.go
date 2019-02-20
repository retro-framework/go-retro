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
	"github.com/retro-framework/go-retro/framework/retro"
)

type HideIdentity struct {
	identity *aggregates.Identity
}

func (cmd *HideIdentity) SetState(agg retro.Aggregate) error {
	if typedAggregate, ok := agg.(*aggregates.Identity); ok {
		cmd.identity = typedAggregate
		return nil
	}
	return errors.New("can't cast")
}

func (cmd *HideIdentity) Apply(ctxt context.Context, w io.Writer, session retro.Session, repo retro.Repo) (retro.CommandResult, error) {

	s := session.(*aggregates.Session)
	if !s.HasIdentity {
		return nil, fmt.Errorf("you haven't authenticated or you don't have an identity")
	}
	if s.IdentityName != cmd.identity.Name() {
		return nil, fmt.Errorf("Identity name on session does not match, this is not your profile")
	}

	json.NewEncoder(w).Encode(cmd.identity)

	return retro.CommandResult{
		cmd.identity: []retro.Event{
			events.SetVisibility{Radius: "private"},
		},
	}, nil
}

func init() {
	commands.Register(&aggregates.Identity{}, &HideIdentity{})
}
