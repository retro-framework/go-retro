package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
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

func New(d types.Depot, r types.ResolveFunc, i types.IDFactory, a types.AggregateManifest) Engine {
	return Engine{
		depot:        d,
		resolver:     r,
		idFactory:    i,
		aggm:         a,
		claimTimeout: 5 * time.Second,
	}
}

type Engine struct {
	depot     types.Depot
	resolver  types.ResolveFunc
	idFactory types.IDFactory

	// TODO: this is a big hammer for finding out which aggregate is registered at "session"
	aggm types.AggregateManifest

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

	// Check we have a repository, aggm and etc, use them to give us
	// back a rehydrated instance of the aggregate, and then we'll
	// call the function on it.
	if a.aggm == nil {
		return "", Error{"agg-manifest-missing", nil, "aggregate manifest not available, please check config."}
	}
	if a.depot == nil {
		return "", Error{"depot-missing", nil, "depot not available, please check config."}
	}
	if a.resolver == nil {
		return "", Error{"resolver-missing", nil, "resolver not available, please check config."}
	}

	var err error

	// Check we have a "session" aggreate in the manifest, else we will struggle
	// from here on out.
	spnSeshAggLookup := opentracing.StartSpan("look up session aggregate", opentracing.ChildOf(spnApply.Context()))
	defer spnSeshAggLookup.Finish()
	seshAgg, err := a.aggm.ForPath("session")
	if err != nil {
		err = Error{"agg-lookup", err, "coult not look up session aggregate in manifest"}
		spnSeshAggLookup.LogKV("event", "error", "error.object", err)
		return "", err
	}

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
		err := a.depot.Rehydrate(ctx, seshAgg, types.PartitionName(sessionPath))
		if err != nil {
			err := Error{"session-lookup", err, "could not look up session"}
			spnRehydrateSesh.LogKV("event", "error", "error.object", err)
			return "", err
		}
		spnRehydrateSesh.Finish()
	}

	spnResolveCmd := opentracing.StartSpan("resolve command", opentracing.ChildOf(spnApply.Context()))
	callable, err := a.resolver(ctx, a.depot, cmd)
	if err != nil {
		return "", errors.Errorf("Couldn't resolve %s", cmd)
	}
	spnResolveCmd.Finish()

	spnApplyCmd := opentracing.StartSpan("apply command", opentracing.ChildOf(spnApply.Context()))
	newEvs, err := callable(ctx, seshAgg, a.depot)
	if err != nil {
		return "", errors.Wrap(err, "error applying command")
	}
	spnApplyCmd.Finish()

	var peek = struct {
		Path string `json:"path"`
	}{}
	err = json.Unmarshal(cmd, &peek)
	if err != nil {
		return "", Error{"peek-path", err, "could not peek into cmd desc to determine path"}
	}

	// TODO: The repo/depot should also store the results
	// of the command and the errors, if any. It should also
	// store the timing of the execution, and respect the
	// Durability of the event. (crit, optimistic)
	spnAppendEvs, ctx := opentracing.StartSpanFromContext(ctx, "store generated events in depot")
	n, err := a.depot.AppendEvs(peek.Path, newEvs)
	if n != len(newEvs) {
		return "", Error{"depot-partial-store", nil, "depot encountered error storing events for aggregate"}
	}
	spnAppendEvs.Finish()

	return "ok", nil
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

	if e.depot.Exists(path) {
		return sid, Error{"guard-unique-session-id", err, "session id was not unique in depot, can't start."}
	}

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
