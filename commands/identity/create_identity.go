package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/retro-framework/go-retro/aggregates"
	"github.com/retro-framework/go-retro/commands"
	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/retro"
)

type CreateArgs struct {
	Name            string `json:"name"`
	PubliclyVisible bool   `json:"publiclyVisible"`
	Avatar          []byte `json:"avatar"`
}

type CreateIdentity struct {
	args CreateArgs
}

func (cmd *CreateIdentity) SetState(agg retro.Aggregate) error {
	return nil
}

func (cmd *CreateIdentity) SetArgs(a retro.CommandArgs) error {
	if typedArgs, ok := a.(*CreateArgs); ok {
		cmd.args = *typedArgs
	} else {
		return fmt.Errorf("can't typecast args")
	}
	return nil
}

func (cmd *CreateIdentity) Apply(ctx context.Context, w io.Writer, session retro.Session, repo retro.Repo) (retro.CommandResult, error) {

	var newIdentity = &aggregates.Identity{}

	var ownEvents = []retro.Event{
		events.SetDisplayName{Name: cmd.args.Name},
	}

	if cmd.args.PubliclyVisible == true {
		ownEvents = append(ownEvents, events.SetVisibility{Radius: "public"})
	}

	if len(cmd.args.Avatar) > 0 {
		ownEvents = append(ownEvents, events.SetAvatar{ImgData: cmd.args.Avatar})
	}

	return retro.CommandResult{
		session:     []retro.Event{&events.AssociateIdentity{Identity: newIdentity}},
		newIdentity: ownEvents,
	}, nil
}

func (cmd *CreateIdentity) Render(ctx context.Context, w io.Writer, session retro.Session, res retro.CommandResult) error {
	json.NewEncoder(w).Encode(res)
	return nil
}

func init() {
	commands.RegisterWithArgs(&aggregates.WidgetsApp{}, &CreateIdentity{}, &CreateArgs{})
}
