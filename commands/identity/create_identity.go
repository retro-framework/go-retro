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

type args struct {
	Name            string `json:"name"`
	PubliclyVisible bool   `json:"publiclyVisible"`
	Avatar          []byte `json:"avatar"`
}

type CreateIdentity struct {
	identity *aggregates.Identity
	args     args
}

// SetState receieves an anonymous Aggregate and must type assert
// it to the correct type (Identity).
func (cmd *CreateIdentity) SetState(agg types.Aggregate) error {
	if typedAggregate, ok := agg.(*aggregates.Identity); ok {
		cmd.identity = typedAggregate
		return nil
	} else {
		return errors.New("can't cast")
	}
}

func (cmd *CreateIdentity) SetArgs(a types.CommandArgs) error {
	if typedArgs, ok := a.(*args); ok {
		cmd.args = *typedArgs
	} else {
		return fmt.Errorf("can't typecast args")
	}
	return nil
}

func (cmd *CreateIdentity) Apply(ctxt context.Context, w io.Writer, session types.Session, repo types.Repository) (types.CommandResult, error) {

	var ownEvents = []types.Event{
		events.SetDisplayName{Name: cmd.args.Name},
	}

	if cmd.args.PubliclyVisible == true {
		ownEvents = append(ownEvents, events.SetVisibility{Radius: "public"})
	}

	if len(cmd.args.Avatar) > 0 {
		ownEvents = append(ownEvents, events.SetAvatar{ImgData: cmd.args.Avatar})
	}

	json.NewEncoder(w).Encode(cmd.identity)

	return types.CommandResult{
		session: []types.Event{
			events.AssociateIdentity{Identity: cmd.identity.Name()},
		},
		cmd.identity: ownEvents,
	}, nil
}

func init() {
	commands.RegisterWithArgs(&aggregates.Identity{}, &CreateIdentity{}, &args{})
}