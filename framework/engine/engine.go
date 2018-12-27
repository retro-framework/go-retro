package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/framework/packing"
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

func New(d types.Depot, r types.ResolveFunc, i types.IDFactory, a types.AggregateManifest, e types.EventManifest) Engine {
	return Engine{
		depot:        d,
		resolver:     r,
		idFactory:    i,
		aggm:         a,
		evm:          e,
		claimTimeout: 5 * time.Second,
	}
}

type Engine struct {
	depot types.Depot

	resolver  types.ResolveFunc
	idFactory types.IDFactory

	// TODO: this is a big hammer for finding out which aggregate is registered at "session"
	aggm types.AggregateManifest
	evm  types.EventManifest

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

	var err error

	// Tracing
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
	commandFn, err := a.resolver(ctx, a.depot, cmd)
	if err != nil {
		return "", errors.Errorf("Couldn't resolve %s", cmd)
	}
	spnResolveCmd.Finish()

	spnApplyCmd := opentracing.StartSpan("apply command", opentracing.ChildOf(spnApply.Context()))
	newEvs, err := commandFn(ctx, seshAgg, a.depot)
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

	if err := a.persistEvs(peek.Path, newEvs); err != nil {
		return "", err // TODO: wrap me
	} else {
		return "ok", nil
	}

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

	// Tracing
	spnResolve, ctx := opentracing.StartSpanFromContext(ctx, "engine.StartSession")
	defer spnResolve.Finish()

	// Generate a session id using the provided factory
	sidStr, err := e.idFactory()
	sid := types.SessionID(sidStr)
	if err != nil {
		return sid, Error{"generate-id-for-session", err, "id factory returned an error when genrating an id"}
	}

	// Tracing
	spnUnmarshal := opentracing.StartSpan("marshal start session command from anon struct", opentracing.ChildOf(spnResolve.Context()))
	path := fmt.Sprintf("session/%s", sid)

	// Guard against reuse of session ids, we could avoid this
	// if we used a cryptographically secure prng in the id factory
	// but users may provide a bad implementation (also, tests.)
	if e.depot.Exists(types.PartitionName(path)) {
		return sid, Error{"guard-unique-session-id", err, "session id was not unique in depot, can't start."}
	}

	// Our cancellation clause, claimTimeout is the wait time we're
	// willing to inflict on our consumer/client waiting for other
	// processes which may have a lock on the aggregate we want.
	claimCtx, cancel := context.WithTimeout(ctx, e.claimTimeout)
	defer cancel()

	// Try and get an exclusive claim on the resource, it will
	// honor the timeout we have already set.
	e.depot.Claim(claimCtx, path)
	defer e.depot.Release(path)

	// Tracing
	b, err := json.Marshal(struct {
		Path    string `json:"path"`
		CmdName string `json:"name"`
	}{path, "Start"})
	if err != nil {
		return sid, Error{"marshal-session-start-cmd", err, "can't marshal session start command to JSON for resolver"}
	}
	spnUnmarshal.SetTag("payload", string(b))
	spnUnmarshal.Finish()

	// A command to start a session is mandatory, even if it's a virtual no-op
	// this looks up the command handler.
	sessionStart, err := e.resolver(ctx, e.depot, b)
	if err != nil {
		return sid, Error{"resolve-session-start-cmd", err, "can't resolve session start command"}
	}

	// Trigger the session handler, and see what events, if any are emitted.
	sessionStartedEvents, err := sessionStart(ctx, nil, e.depot)
	if err != nil {
		return sid, Error{"execute-session-start-cmd", err, "error calling session start command"}
	}

	return sid, e.persistEvs(path, sessionStartedEvents)

	// Tracing
	// spnAppendEvs, ctx := opentracing.StartSpanFromContext(ctx, "store generated events in depot")
	// // n, err := e.depot.AppendEvs(path, evs)
	// // if n != len(evs) {
	// // 	return sid, Error{"depot-partial-store", err, "depot encountered error storing events for aggregate"}
	// // }
	// spnAppendEvs.Finish()

	// return sid, nil
}

// TODO: fix all the error messages (or Error types, etc, who knows.)
func (e *Engine) persistEvs(path string, evs []types.Event) error {

	var (
		jp          = packing.NewJSONPacker()
		affix       = packing.Affix{}
		packedeObjs []types.HashedObject
	)

	for _, ev := range evs {

		name, err := e.evm.KeyFor(ev)
		if err != nil {
			return Error{"persist-events-from-session-start", err, "error looking up event"}
		}

		packedEv, err := jp.PackEvent(name, ev)
		if err != nil {
			return Error{"persist-events-from-session-start", err, "error packing event"}
		}

		packedeObjs = append(packedeObjs, packedEv)
		affix[types.PartitionName(path)] = append(affix[types.PartitionName(path)], packedEv.Hash())

	}

	packedAffix, err := jp.PackAffix(affix)
	if err != nil {
		return Error{"persist-evs", err, "error packing affix in NewSimpleStub: %s"}
	}
	packedeObjs = append(packedeObjs, packedAffix)

	// TODO: packed checkpoint needs a parent, else we will make orphan stuff
	checkpoint := packing.Checkpoint{
		AffixHash:   packedAffix.Hash(),
		CommandDesc: []byte(`{"stub":"article"}`),
		Fields:      map[string]string{"session": "hello world"},
	}

	packedCheckpoint, err := jp.PackCheckpoint(checkpoint)
	if err != nil {
		return Error{"persist-evs", err, "error packing checkpoint in NewSimpleStub: %s"}
	}
	packedeObjs = append(packedeObjs, packedCheckpoint)

	if err := e.depot.StorePacked(packedeObjs...); err != nil {
		return Error{"persist-evs", err, "error writing packedAffix to odb in NewSimpleStub"}
	}

	if err := e.depot.MoveHeadPointer(nil, packedCheckpoint.Hash()); err != nil {
		return Error{"persist-evs", err, "moving head pointer"}
	}

	return nil

}
