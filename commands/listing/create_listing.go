package listing

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/retro-framework/go-retro/aggregates"
	"github.com/retro-framework/go-retro/commands"
	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/retro"
)

type Args struct {
	Name       string   `json:"name"`
	Desc       string   `json:"desc"`
	PublishNow bool     `json:"publishNow"`
	StartPrice uint16   `json:"startPrice"`
	Images     [][]byte `json:"images"`
}

type CreateListing struct {
	wa   *aggregates.WidgetsApp
	args Args
}

func (cmd *CreateListing) SetState(agg retro.Aggregate) error {
	if typedAggregate, ok := agg.(*aggregates.WidgetsApp); ok {
		cmd.wa = typedAggregate
		return nil
	}
	return errors.New("can't cast")
}

func (cmd *CreateListing) SetArgs(a retro.CommandArgs) error {
	if typedArgs, ok := a.(*Args); ok {
		cmd.args = *typedArgs
	} else {
		return fmt.Errorf("can't typecast args")
	}
	return nil
}

func (cmd *CreateListing) Apply(ctx context.Context, w io.Writer, session retro.Session, repo retro.Repo) (retro.CommandResult, error) {

	if s, ok := session.(*aggregates.Session); ok {
		if !s.HasIdentity {
			return nil, fmt.Errorf("can't create listing on anon session")
		}
		// TODO: maybe recommend a command to issue first?
	} else {
		return nil, fmt.Errorf("can't cast session, world is broken")
	}

	if !cmd.wa.CreationOfListingsAllowed() {
		return nil, fmt.Errorf("creation of listings is forbidden")
	}

	var listing = &aggregates.Listing{}

	var ownEvents = []retro.Event{
		events.SetDisplayName{Name: cmd.args.Name},
		events.SetDescription{Desc: cmd.args.Desc},
	}

	for _, img := range cmd.args.Images {
		ownEvents = append(ownEvents, events.CreateListingImage{img})
	}

	if cmd.args.PublishNow {
		ownEvents = append(ownEvents, events.SetVisibility{Radius: "public"})
	}

	return retro.CommandResult{listing: ownEvents}, nil
}

// Render looks for the first aggregates.Listing in the
// results, and prints its name so the caller can use
// that when setting up the redirections/etc.
func (cmd *CreateListing) Render(ctx context.Context, w io.Writer, session retro.Session, res retro.CommandResult) error {
	for k := range res {
		if _, ok := k.(*aggregates.Listing); ok {
			fmt.Fprint(w, k.Name())
			return nil
		}
	}
	return nil
}

func init() {
	commands.RegisterWithArgs(&aggregates.WidgetsApp{}, &CreateListing{}, &Args{})
}
