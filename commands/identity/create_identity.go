package identity

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

func (cmd *CreateIdentity) Apply(ctx context.Context, w io.Writer, session types.Session, repo types.Repository) (types.CommandResult, error) {

	// if repo.Exists(ctx, cmd.identity.Name()) {
	// 	return nil, fmt.Errorf("identity already exists with name %q", cmd.identity.Name())
	// }

	// var s = session.(*aggregates.Session)
	// if s.HasIdentity && s.IdentityName == cmd.identity.Name() {
	// 	return nil, fmt.Errorf("session %s is already associated with an identity named %s (%t)", s.Name(), cmd.identity.Name(), exists)
	// }

	var ownEvents = []types.Event{
		events.SetDisplayName{Name: cmd.args.Name},
	}

	if cmd.args.PubliclyVisible == true {
		ownEvents = append(ownEvents, events.SetVisibility{Radius: "public"})
	}

	if len(cmd.args.Avatar) > 0 {
		ownEvents = append(ownEvents, events.SetAvatar{ImgData: cmd.args.Avatar})
	}

	return types.CommandResult{
		session: []types.Event{
			&events.AssociateIdentity{Identity: cmd.identity},
		},
		&aggregates.Identity{}: ownEvents,
	}, nil
}

func init() {
	commands.RegisterWithArgs(&aggregates.WidgetsApp{}, &CreateIdentity{}, &args{})
}
