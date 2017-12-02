package main

import (
	"context"
	"path/filepath"
	"time"

	"github.com/leehambley/ls-cms/aggregates"
	"github.com/leehambley/ls-cms/framework/types"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

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
func (a *Engine) Apply(ctxt context.Context, sid types.SessionID, cmd types.CommandDesc) (string, error) {

	var sp opentracing.Span = opentracing.StartSpan("app/apply")
	defer sp.Finish()

	var (
		sesh = &aggregates.Session{}
		err  error
	)

	// If a session ID was provided, look it up in the repository.  preliminary
	// checks on the session could/should be done here to avoid every command
	// having to repeat the is/not valid checks (although some commands such as
	// changing passwords may want to do more thorough checks on sessions such as
	// matching session IP address to the current client address from the ctxt?).
	if sid != "" {
		sessionPath := filepath.Join("sessions", string(sid))
		err := a.depot.Rehydrate(sesh, sessionPath)
		if err != nil {
			return "", errors.Wrap(err, "could not look up session")
		} else {
			_ = sesh
		}
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

	callable, err := a.resolver(a.depot, cmd)
	if err != nil {
		return "", errors.Errorf("Couldn't resolve %s", cmd)
	}

	start := time.Now()

	_, err = callable(ctxt, sesh, a.depot)
	if err != nil {
		return "", errors.Wrap(err, "error from downstream")
	}

	_ = time.Since(start)

	return "", nil
}

func (a *Engine) StartSession(types.SessionParams) {

}
