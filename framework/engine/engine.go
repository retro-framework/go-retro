package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"time"

	"github.com/gobuffalo/flect"
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

func New(d types.Depot, r types.Repository, resolver types.Resolver, i types.IDFactory, c types.Clock, a types.AggregateManifest, e types.EventManifest) Engine {
	return Engine{
		depot:        d,
		repository:   r,
		resolver:     resolver,
		idFactory:    i,
		clock:        c,
		aggm:         a,
		evm:          e,
		claimTimeout: 5 * time.Second,
	}
}

type Engine struct {
	depot      types.Depot
	repository types.Repository

	resolver  types.Resolver
	idFactory types.IDFactory
	clock     types.Clock

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
//
// Apply cannot write to an empty Depot, it will propagate depot.ErrUnknownRef
// as a real error. SessionStart will allow writing to an empty depot however
func (e *Engine) Apply(ctx context.Context, w io.Writer, sid types.SessionID, cmd []byte) (string, error) {

	var err error

	// Tracing
	spnApply, ctx := opentracing.StartSpanFromContext(ctx, "engine.Apply")
	spnApply.SetTag("payload", string(cmd))
	defer spnApply.Finish()

	// Check we have a repository, aggm and etc, use them to give us
	// back a rehydrated instance of the aggregate, and then we'll
	// call the function on it.
	if e.aggm == nil {
		return "", Error{"agg-manifest-missing", nil, "aggregate manifest not available, please check config."}
	}
	if e.depot == nil {
		return "", Error{"depot-missing", nil, "depot not available, please check config."}
	}
	if e.resolver == nil {
		return "", Error{"resolver-missing", nil, "resolver not available, please check config."}
	}

	headPtr, err := e.depot.HeadPointer(ctx)
	if headPtr == nil && err == nil {
		return "", Error{Op: "get-head-pointer", Msg: "depot is empty, start a session first"}
	}

	if err != nil {
		err = Error{"get-head-pointer", err, "could not get head pointer from depot"}
		spnApply.SetTag("error", err.Error())
		return "", err
	}

	if headPtr != nil {
		spnApply.SetTag("head pointer", headPtr.String())
	}

	// Check we have a "session" aggreate in the manifest, else we will struggle
	// from here on out.
	spnSeshAggLookup := opentracing.StartSpan("look up session aggregate", opentracing.ChildOf(spnApply.Context()))
	defer spnSeshAggLookup.Finish()
	seshAgg, err := e.aggm.ForPath("session")
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
		if err := seshAgg.SetName(types.PartitionName(sessionPath)); err != nil {
			err := Error{"session-lookup", err, "aggregate name did not persist"}
			return "", err
		}
		err := e.repository.Rehydrate(ctx, seshAgg, types.PartitionName(sessionPath))
		if err != nil {
			err := Error{"session-lookup", err, "could not look up session"}
			spnRehydrateSesh.LogKV("event", "error", "error.object", err)
			return "", err
		}
		spnRehydrateSesh.Finish()
	}

	spnResolveCmd := opentracing.StartSpan("resolve command", opentracing.ChildOf(spnApply.Context()))
	command, err := e.resolver.Resolve(ctx, e.repository, cmd)
	if err != nil {
		return "", errors.Errorf("Couldn't resolve %s (%s)", cmd, err)
	}
	spnResolveCmd.Finish()

	spnApplyCmd := opentracing.StartSpan("apply command", opentracing.ChildOf(spnApply.Context()))
	newEvs, err := command.Apply(ctx, w, seshAgg, e.repository)
	if err != nil {
		return "", errors.Wrap(err, "error applying command")
	}
	spnApplyCmd.Finish()

	if err := e.persistEvs(ctx, sid, cmd, headPtr, newEvs); err != nil {
		return "", err // TODO: wrap me
	}

	if commandWithRenderFn, hasRenderFn := command.(types.CommandWithRenderFn); hasRenderFn {
		commandWithRenderFn.Render(ctx, w, seshAgg, newEvs)
	}

	return "ok", nil
}

func dumpCommandResult(w io.Writer, cr types.CommandResult) {
	fmt.Fprint(w, "Command Result Dump:\n")
	var m = make(map[types.PartitionName][]types.Event)
	for k, v := range cr {
		m[k.Name()] = v
	}
	fmt.Fprintf(w, "%#v\n", m)
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
//
// If the Depot is completely empty a session must be started first before
// using Apply. Apply will not gracefully handle seeing a depot.ErrUnknownRef
// but SessionStart will.
func (e *Engine) StartSession(ctx context.Context) (types.SessionID, error) {

	// Tracing
	spnStartSession, ctx := opentracing.StartSpanFromContext(ctx, "engine.StartSession")
	defer spnStartSession.Finish()

	// Generate a session id using the provided factory
	sidStr, err := e.idFactory()
	if err != nil {
		return "", Error{"generate-id-for-session", err, "id factory returned an error when genrating an id"}
	}
	var sid = types.SessionID(sidStr)

	// Guard against reuse of session ids, we could avoid this
	// if we used a cryptographically secure prng in the id factory
	// but users may provide a bad implementation (also, tests.)
	var path = fmt.Sprintf("session/%s", sid)
	if e.repository.Exists(ctx, types.PartitionName(path)) {
		return sid, Error{Op: "guard-unique-session-id", Msg: fmt.Sprintf("session id %q was not unique in depot, can't start.", path)}
	}

	headPtr, err := e.depot.HeadPointer(ctx)
	if err != nil {
		err = Error{"get-head-pointer", err, "could not get head pointer from depot"}
		spnStartSession.SetTag("error", err)
		return "", err
	}
	if headPtr != nil {
		spnStartSession.SetTag("head pointer", headPtr.String())
	}

	// Tracing
	spnUnmarshal := opentracing.StartSpan("marshal start session command from anon struct", opentracing.ChildOf(spnStartSession.Context()))

	// Our cancellation clause, claimTimeout is the wait time we're
	// willing to inflict on our consumer/client waiting for other
	// processes which may have a lock on the aggregate we want.
	claimCtx, cancel := context.WithTimeout(ctx, e.claimTimeout)
	defer cancel()

	// Try and get an exclusive claim on the resource, it will
	// honor the timeout we have already set.
	e.repository.Claim(claimCtx, path)
	defer e.repository.Release(path)

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
	sessionStartCmd, err := e.resolver.Resolve(ctx, e.repository, b)
	if err != nil {
		return sid, Error{"resolve-session-start-cmd", err, "can't resolve session start command"}
	}

	// Trigger the session handler, and see what events, if any are emitted.
	var buf bytes.Buffer
	sessionStartedEvents, err := sessionStartCmd.Apply(ctx, &buf, nil, e.repository)
	if err != nil {
		return sid, Error{"execute-session-start-cmd", err, "error calling session start command"}
	}

	return sid, e.persistEvs(ctx, sid, b, headPtr, sessionStartedEvents)

	// Tracing
	// spnAppendEvs, ctx := opentracing.StartSpanFromContext(ctx, "store generated events in depot")
	// // n, err := e.depot.AppendEvs(path, evs)
	// // if n != len(evs) {
	// // 	return sid, Error{"depot-partial-store", err, "depot encountered error storing events for aggregate"}
	// // }
	// spnAppendEvs.Finish()
}

// guardHasID takes an aggregate, and checks if it has a name, if so then
// it will pass-thru else it will leverage the ID generating function
// to make one before persisting.
//
// TODO: could also check if an aggregate has a bogus name that doesn't
// match its type.
func (e *Engine) guardHasID(a types.Aggregate) (types.PartitionName, error) {

	var (
		name   types.PartitionName = a.Name()
		toType                     = func(t types.Aggregate) reflect.Type {
			// TODO this code is copy-pasted from the aggregates.go manifest
			// it should be consolidated
			var v = reflect.ValueOf(t)
			if reflect.Ptr == v.Kind() || reflect.Interface == v.Kind() {
				v = v.Elem()
			}
			return v.Type()
		}
		err error
	)
	if len(name) == 0 {
		var id, err = e.idFactory()
		if err != nil {
			return "", err
		}
		var n = filepath.Join(flect.Underscore(toType(a).Name()), id)
		name = types.PartitionName(n)
	}
	a.SetName(name)
	return name, err
}

// TODO: Fix all the error messages (or Error types, etc, who knows.)
//
// TODO: extracting writing the checkpoint from here would be a good separation
// of concerns.
func (e *Engine) persistEvs(ctx context.Context, sid types.SessionID, cmdDesc []byte, head types.Hash, cmdRes types.CommandResult) error {

	var (
		jp          = packing.NewJSONPacker()
		affix       = packing.Affix{}
		packedeObjs []types.HashedObject
	)

	currentHead, err := e.depot.HeadPointer(ctx)
	if err != nil {
		return err // TODO: Wrap me & test (?)
	}

	// parentHashes can be complicated to infer, so pull that logic out
	// here to a variable and a set of conditional statements to make
	// constructing the packing.Checkpoint easier below.
	var parentHashes []types.Hash

	// There are possible cases here, that currentHead and head are both, or indiviudally
	// nil. This can only happen incase the ctx carries a branch name which has no history
	// (refs) or the depot is just empty (actually that's basically the same case)
	if head != nil && currentHead != nil {
		// TODO handle this case gracefully. It ought to be possible to check if any of the objects
		// we had in our cmdRes conflict with changes that landed between head and currentHead. This
		// ventures into the territory of merge conflict resolution/etc. Safer might be to return
		// a sentinel error which the Engine can use to retry the entire transaction on top of the
		// new head.
		//
		// This case should be safe as the early returns incase either the left or right hand sides
		// are nil will ensure we never memory fault here.
		if currentHead.String() != head.String() {
			return fmt.Errorf("concurrent write, head pointer moved since we claimed it")
		}
		parentHashes = append(parentHashes, head)
	}

	if head == nil && currentHead != nil {
		// This case means head was nil when Apply was called, but by the time we got to
		// here someo
		return fmt.Errorf("concurrent write, branch created since operation started")
	}

	if head != nil && currentHead == nil {
		// This means someone has deleted our branch since we started, and we're about
		// to recreate it if we continue.
		return fmt.Errorf("concurrent write, branch deleted since operation started")
	}

	for agg, evs := range cmdRes {
		aggPath, err := e.guardHasID(agg)
		if err != nil {
			return Error{"persist-evs", err, "could not generate id for anonymous aggregate"}
		}
		for _, ev := range evs {
			name, err := e.evm.KeyFor(ev)
			if err != nil {
				return Error{"persist-evs", err, "error looking up event"}
			}
			packedEv, err := jp.PackEvent(name, ev)
			if err != nil {
				return Error{"persist-evs", err, "error packing event"}
			}
			packedeObjs = append(packedeObjs, packedEv)
			affix[aggPath] = append(affix[aggPath], packedEv.Hash())
		}
	}

	packedAffix, err := jp.PackAffix(affix)
	if err != nil {
		return Error{"persist-evs", err, "error packing affix in NewSimpleStub: %s"}
	}
	packedeObjs = append(packedeObjs, packedAffix)

	checkpoint := packing.Checkpoint{
		AffixHash:   packedAffix.Hash(),
		CommandDesc: cmdDesc,
		Fields: map[string]string{
			"session": string(sid),
			"date":    e.clock.Now().Format(time.RFC3339),
		},
		ParentHashes: parentHashes,
	}

	if _, err := checkpoint.HasErrors(); len(err) > 0 {
		// TODO: HasErrors can return a bunch of errors
		// we should do something smarter here.
		return Error{"persist-evs", err[0], "error with the checkpoint validations"}
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
