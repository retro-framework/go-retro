package main

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/aggregates"
	"github.com/retro-framework/go-retro/framework/types"
)

type Error struct {
	Op  string
	Err error
	Msg string
}

func (e Error) Error() string {
	return fmt.Sprintf("engine: op: %q err: %q msg: %q", e.Op, e.Err, e.Msg)
}

type Engine struct {
	log      types.Logger
	tracer   opentracing.Tracer
	depot    types.Depot
	resolver types.ResolveFunc
}

// Apply takes a command and uses a Resolver to determine which aggregate
// to dispatch it to. Commands without an aggregate are dispatched to the
// root object (_). Commands are applied through a Repository which is used
// to load and save aggregates. Commands, and their resulting events are
// logged to the event store. Aggregates should not store data that is not
// relevant for business decisions Aggregates should not store data that is
// not relevant for business decisions. Aggregates should also not
// "contain" unique IDs, they should be assigned from the outside.
//
// Apply needs to lookup the method from the description and apply it.
// There is presently no way (should there be?) to construct a command "by
// hand" it must be serializable to account for the repo rehydrating the
// aggregate.
func (a *Engine) Apply(ctx context.Context, sid types.SessionID, cmd []byte) (string, error) {

	spnApply, ctx := opentracing.StartSpanFromContext(ctx, "engine.Apply")
	spnApply.SetTag("payload", string(cmd))
	defer spnApply.Finish()

	var (
		sesh = &aggregates.Session{}
		err  error
	)

	// If a session ID was provided, look it up in the repository.  preliminary
	// checks on the session could/should be done here to avoid every command
	// having to repeat the is/not valid checks (although some commands such as
	// changing passwords may want to do more thorough checks on sessions such as
	// matching session IP address to the current client address from the ctxt?).
	if sid == "" {
		spnApply.LogEvent("no session found")
	} else {
		sessionPath := filepath.Join("sessions", string(sid))
		spnRehydrateSesh := opentracing.StartSpan("rehydrating session", opentracing.ChildOf(spnApply.Context()))
		defer spnRehydrateSesh.Finish()
		err := a.depot.Rehydrate(ctx, sesh, sessionPath)
		if err != nil {
			err := Error{"session-lookup", err, "could not look up session"}
			spnRehydrateSesh.LogKV("event", "error", "error.object", err)
			return "", err
		}
		spnRehydrateSesh.Finish()
	}

	// Check we have a repository, and use the repo and the resolver to
	// give us back a rehydrated instance of the aggregate, and then we'll
	// call the function on it.
	if a.depot == nil {
		return "", errors.New("repository not defined, please check config.")
	}
	if a.resolver == nil {
		return "", errors.New("resolver not defined, please check config")
	}

	callable, err := a.resolver(ctx, a.depot, cmd)
	if err != nil {
		return "", errors.Errorf("Couldn't resolve %s", cmd)
	}

	start := time.Now()

	_, err = callable(ctx, sesh, a.depot)
	if err != nil {
		return "", errors.Wrap(err, "error from downstream")
	}

	_ = time.Since(start)

	return "", nil
}

func (a *Engine) StartSession(types.SessionParams) {

}
