package engine

import (
	"context"
	"encoding/json"
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

func NewEngine(d types.Depot, r types.ResolveFunc, i types.IDFactory) Engine {
	return Engine{d, r, i, 5 * time.Second}
}

func New(d types.Depot, r types.ResolveFunc, i types.IDFactory) Engine {
	return NewEngine(d, r, i)
}

type Engine struct {
	depot     types.Depot
	resolver  types.ResolveFunc
	idFactory types.IDFactory

	claimTimeout time.Duration
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
		sessionPath := filepath.Join("session", string(sid))
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

	_, err = callable(ctx, sesh, a.depot)
	if err != nil {
		return "", errors.Wrap(err, "error from downstream")
	}

	return "", nil
}

// StartSession will attempt to summon a Session aggregate into existence by
// calling it's "Start" command. If no session aggregate is registered and no
// "start" command for it exists an error will be raised.
//
// The ID is generated internally using the types.IDFactory given to the
// constructor. This function can be tied into a ticket server, or a simple
// central sequential store or random hex string generator function at will.
//
// If error is non-nil the SID is not usable. The session ID is returned to
// facilitate correlating logs with failed session start commands.
func (e *Engine) StartSession(ctx context.Context) (types.SessionID, error) {

	spnResolve, ctx := opentracing.StartSpanFromContext(ctx, "engine.StartSession")
	defer spnResolve.Finish()

	sidStr, err := e.idFactory()
	sid := types.SessionID(sidStr)
	if err != nil {
		return sid, Error{"generate-id-for-session", err, "id factory returned an error when genrating an id"}
	}

	spnUnmarshal := opentracing.StartSpan("marshal start session command from anon struct", opentracing.ChildOf(spnResolve.Context()))
	path := fmt.Sprintf("session/%s", sid)

	claimCtx, cancel := context.WithTimeout(ctx, e.claimTimeout)
	defer cancel()

	e.depot.Claim(claimCtx, path)
	defer e.depot.Release(path)

	b, err := json.Marshal(struct {
		Path    string `json:"path"`
		CmdName string `json:"name"`
	}{path, "Start"})
	if err != nil {
		return sid, Error{"marshal-session-start-cmd", err, "can't marshal session start command to JSON for resolver"}
	}
	spnUnmarshal.SetTag("payload", string(b))
	spnUnmarshal.Finish()

	//↓ ResolveFunc needs to return the aggregate really, else I've no handle on it, and I can't "lock" it in the depot
	sessionStart, err := e.resolver(ctx, e.depot, b)
	if err != nil {
		return sid, Error{"resolve-session-start-cmd", err, "can't resolve session start command, no aggregate or command registered"}
	}

	evs, err := sessionStart(ctx, nil, e.depot)
	if err != nil {
		return sid, Error{"execute-session-start-cmd", err, "can't call session start command, an error was returned"}
	}

	spnAppendEvs, ctx := opentracing.StartSpanFromContext(ctx, "store generated events in depot")
	n, err := e.depot.AppendEvs(path, evs)
	if n != len(evs) {
		return sid, Error{"depot-partial-store", err, "depot encountered error storing events for aggregate"}
	}
	spnAppendEvs.Finish()

	return sid, nil
}
