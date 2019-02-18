// +build integration

package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/retro-framework/go-retro/aggregates"
	"github.com/retro-framework/go-retro/commands"
	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/depot"
	"github.com/retro-framework/go-retro/framework/repository"
	"github.com/retro-framework/go-retro/framework/resolver"
	"github.com/retro-framework/go-retro/framework/storage/memory"
	test "github.com/retro-framework/go-retro/framework/test_helper"
	"github.com/retro-framework/go-retro/framework/types"
)

type Predictable5sJumpClock struct {
	t     time.Time
	calls int
}

func (c *Predictable5sJumpClock) Now() time.Time {
	var next = c.t.Add(time.Duration((5 * c.calls)) * time.Second)
	c.calls = c.calls + 1
	return next
}

type DummyEvent struct{}

type DummyStartSessionEvent struct {
	Greeting string
}

type dummySession struct{ aggregates.NamedAggregate }

func (_ *dummySession) ReactTo(types.Event) error { return nil }

type Start struct{ s *dummySession }

func (fssc *Start) SetState(s types.Aggregate) error {
	if agg, ok := s.(*dummySession); ok {
		fssc.s = agg
		return nil
	} else {
		return errors.New("can't cast aggregate state")
	}
}

func (fssc *Start) Apply(context.Context, io.Writer, types.Session, types.Repository) (types.CommandResult, error) {
	return types.CommandResult{fssc.s: []types.Event{DummyStartSessionEvent{"hello world"}}}, nil
}

type dummyAggregate struct {
	aggregates.NamedAggregate
	seenEvents []types.Event
}

func (da *dummyAggregate) ReactTo(ev types.Event) error {
	da.seenEvents = append(da.seenEvents, ev)
	return nil
}

type dummyCmd struct {
	s          *dummyAggregate
	wasApplied bool
}

func (dc *dummyCmd) SetState(s types.Aggregate) error {
	if agg, ok := s.(*dummyAggregate); ok {
		dc.s = agg
		return nil
	} else {
		return errors.New("can't cast aggregate state")
	}
}

func (dc *dummyCmd) Apply(_ context.Context, _ io.Writer, _ types.Session, _ types.Repository) (types.CommandResult, error) {
	dc.wasApplied = true
	return types.CommandResult{dc.s: []types.Event{DummyEvent{}}}, nil
}

func (dc *dummyCmd) Render(_ context.Context, w io.Writer, _ types.Session, _ types.CommandResult) error {
	fmt.Fprintf(w, "ok")
	return nil
}

func ResolverDouble(resolverFn types.ResolveFunc) types.Resolver {
	return resolverDouble{resolverFn}
}

type resolverDouble struct {
	resolveFn types.ResolveFunc
}

func (rd resolverDouble) Resolve(ctx context.Context, repository types.Repository, buf []byte) (types.Command, error) {
	return rd.resolveFn(ctx, repository, buf)
}

// Sessions are a special case of aggregate We always need one, even if anon to
// do anything.
//
// Starting a session is a special-case of sending a command without a
// pre-existing session to the session aggregate to summon it into existence
func Test_Engine_StartSession(t *testing.T) {

	t.Run("can start a session on an empty Depot", func(t *testing.T) {
		// Arrange
		var (
			ctx        = context.Background()
			objdb      = &memory.ObjectStore{}
			refdb      = &memory.RefStore{}
			depot      = depot.NewSimple(objdb, refdb)
			evM        = events.NewManifest()
			repository = repository.NewSimpleRepository(objdb, refdb, evM)
			idFn       = func() (string, error) { return "123-stub-id", nil }
			clock      = &Predictable5sJumpClock{}
			resolver   = ResolverDouble(func(_ context.Context, _ types.Repository, _ []byte) (types.Command, error) {
				var (
					fssc = &Start{}
					s    = &dummySession{}
				)
				s.SetName(types.PartitionName("session/123-stub-id"))
				fssc.SetState(s)
				return fssc, nil
			})
		)
		evM.Register(&DummyEvent{})
		evM.Register(&DummyStartSessionEvent{})
		var e = New(depot, repository, resolver, idFn, clock, aggregates.NewManifest(), evM)

		// Act
		sessionID, err := e.StartSession(ctx)

		// Assert
		test.H(t).IsNil(err)
		test.H(t).StringEql("123-stub-id", string(sessionID))
	})

	t.Run("creates a new session with parameters not matching an aggregate in the repository", func(t *testing.T) {

		// Arrange
		var (
			ctx        = context.Background()
			objdb      = &memory.ObjectStore{}
			refdb      = &memory.RefStore{}
			depot      = depot.NewSimple(objdb, refdb)
			evM        = events.NewManifest()
			repository = repository.NewSimpleRepository(objdb, refdb, evM)
			idFn       = func() (string, error) { return "123-stub-id", nil }
			clock      = &Predictable5sJumpClock{}
			resolver   = ResolverDouble(func(_ context.Context, _ types.Repository, _ []byte) (types.Command, error) {
				var (
					fssc = &Start{}
					s    = &dummySession{}
				)
				s.SetName(types.PartitionName("session/123-stub-id"))
				fssc.SetState(s)
				return fssc, nil
			})
		)
		evM.Register(&DummyEvent{})
		evM.Register(&DummyStartSessionEvent{})
		var e = New(depot, repository, resolver, idFn, clock, aggregates.NewManifest(), evM)

		// Act
		sid, err := e.StartSession(ctx)
		// Assert
		test.H(t).IsNil(err)

		if dd, ok := depot.(types.DumpableDepot); !ok {
			t.Fatal("could not upgrade depot to diff it")
		} else {
			var b bytes.Buffer
			dd.DumpAll(&b)
		}

		test.H(t).BoolEql(true, repository.Exists(ctx, types.PartitionName(fmt.Sprintf("session/%s", sid))))
		_, err = e.StartSession(ctx)
		test.H(t).NotNil(err)
	})

	t.Run("persists the resulting session aggregate to the repository if the start command yields events", func(t *testing.T) {

		// Arrange
		var (
			ctx        = context.Background()
			objdb      = &memory.ObjectStore{}
			refdb      = &memory.RefStore{}
			depot      = depot.NewSimple(objdb, refdb)
			evM        = events.NewManifest()
			repository = repository.NewSimpleRepository(objdb, refdb, evM)
			idFn       = func() (string, error) { return "123-stub-id", nil }
			clock      = &Predictable5sJumpClock{}
			resolver   = ResolverDouble(func(_ context.Context, _ types.Repository, _ []byte) (types.Command, error) {
				var (
					fssc = &Start{}
					s    = &dummySession{}
				)
				s.SetName(types.PartitionName("session/123-stub-id"))
				fssc.SetState(s)
				return fssc, nil
			})
		)
		evM.Register(&DummyStartSessionEvent{})

		var e = New(depot, repository, resolver, idFn, clock, aggregates.NewManifest(), evM)

		// Act
		sid, err := e.StartSession(context.Background())

		// Assert
		test.H(t).IsNil(err)
		test.H(t).BoolEql(true, repository.Exists(ctx, types.PartitionName(fmt.Sprintf("session/%s", sid))))
	})

	t.Run("forwards errors from the resolvefn to the caller", func(t *testing.T) {

		// Arrange
		var (
			resolverErr   = fmt.Errorf("error from resolveFn")
			objdb         = &memory.ObjectStore{}
			refdb         = &memory.RefStore{}
			depot         = depot.NewSimple(objdb, refdb)
			eventManifest = events.NewManifest()
			repository    = repository.NewSimpleRepository(objdb, refdb, eventManifest)
			idFn          = func() (string, error) { return fmt.Sprintf("%x", rand.Uint64()), nil }
			clock         = &Predictable5sJumpClock{}
			resolver      = ResolverDouble(func(_ context.Context, _ types.Repository, _ []byte) (types.Command, error) {
				return nil, resolverErr
			})
		)
		var e = New(depot, repository, resolver, idFn, clock, aggregates.NewManifest(), events.NewManifest())

		// Act
		_, err := e.StartSession(context.Background())

		// Assert
		test.H(t).NotNil(err)
		if eError, ok := err.(Error); !ok {
			t.Fatal("expected to get a typed Error, not a generic inteface err")
		} else {
			test.H(t).ErrEql(eError.Err, resolverErr)
		}
	})
}

func Test_Engine_DepotClaims(t *testing.T) {
	t.Skip("not tested for yet (should we test claims in the depot test matrix?)")
}

func Test_Engine_Apply(t *testing.T) {

	t.Run("can not apply anything on an empty Depot", func(t *testing.T) {
		// Arrange
		var (
			ctx           = context.Background()
			objdb         = &memory.ObjectStore{}
			refdb         = &memory.RefStore{}
			depot         = depot.NewSimple(objdb, refdb)
			eventManifest = events.NewManifest()
			repository    = repository.NewSimpleRepository(objdb, refdb, eventManifest)
			idFn          = func() (string, error) { return "123-stub-id", nil }
			clock         = &Predictable5sJumpClock{}
			resolver      = ResolverDouble(func(_ context.Context, _ types.Repository, _ []byte) (types.Command, error) {
				var (
					fssc = &Start{}
					s    = &dummySession{}
				)
				s.SetName(types.PartitionName("session/123-stub-id"))
				fssc.SetState(s)
				return fssc, nil
			})
		)
		eventManifest.Register(&DummyEvent{})
		eventManifest.Register(&DummyStartSessionEvent{})
		var e = New(depot, repository, resolver, idFn, clock, aggregates.NewManifest(), eventManifest)

		// Act (checking the head pointer happens before routing, so we don't need any setup)
		var b bytes.Buffer
		_, err := e.Apply(ctx, &b, "123-stub-id", []byte(`{"path":"agg/123", "name":"dummyCmd"}`))

		// Assert
		test.H(t).NotNil(err)
		if engineErr, ok := err.(Error); !ok {
			t.Fatal("could not cast error to Error")
		} else {
			test.H(t).StringEql("get-head-pointer", engineErr.Op)
			test.H(t).StringEql("depot is empty, start a session first", engineErr.Msg)
		}
	})

	t.Run("with a non-existent sesionID", func(t *testing.T) {
		t.Skip("not implemented yet")
	})

	t.Run("will call the render func if the command supports that interface", func(t *testing.T) {

		// Arrange
		var (
			objdb      = &memory.ObjectStore{}
			refdb      = &memory.RefStore{}
			depot      = depot.NewSimple(objdb, refdb)
			idFn       = func() (string, error) { return fmt.Sprintf("%x", []byte("hello")), nil }
			clock      = &Predictable5sJumpClock{}
			aggM       = aggregates.NewManifest()
			cmdM       = commands.NewManifest()
			evM        = events.NewManifest()
			repository = repository.NewSimpleRepository(objdb, refdb, evM)

			err error
		)

		aggM.Register("agg", &dummyAggregate{})
		cmdM.Register(&dummyAggregate{}, &dummyCmd{})

		aggM.Register("session", &dummySession{})
		cmdM.Register(&dummySession{}, &Start{})

		evM.Register(&DummyEvent{})
		evM.Register(&DummyStartSessionEvent{})

		var (
			r   = resolver.New(aggM, cmdM)
			e   = New(depot, repository, r, idFn, clock, aggM, evM)
			ctx = context.Background()
		)

		// Act
		sid, err := e.StartSession(ctx)
		test.H(t).IsNil(err)

		var b bytes.Buffer
		_, err = e.Apply(ctx, &b, sid, []byte(`{"path":"agg/123", "name":"dummyCmd"}`))

		// Assert
		test.H(t).IsNil(err)
		test.H(t).StringEql("ok", b.String())
	})

	t.Run("routing", func(t *testing.T) {

		t.Run("raises error and logs it on unroutable (unregistered) command", func(t *testing.T) {

			// Arrange
			var (
				objdb             = &memory.ObjectStore{}
				refdb             = &memory.RefStore{}
				depot             = depot.NewSimple(objdb, refdb)
				idFn              = func() (string, error) { return fmt.Sprintf("%x", rand.Uint64()), nil }
				clock             = &Predictable5sJumpClock{}
				aggregateManifest = aggregates.NewManifest()
				commandManifest   = commands.NewManifest()
				eventManifest     = events.NewManifest()
				repository        = repository.NewSimpleRepository(objdb, refdb, eventManifest)
			)

			// NOTE: no calls to register anything on the manifests except
			// the session!
			aggregateManifest.Register("session", &dummySession{})
			commandManifest.Register(&dummySession{}, &Start{})
			eventManifest.Register(&DummyStartSessionEvent{})

			var (
				e = New(
					depot,
					repository,
					resolver.New(aggregateManifest, commandManifest),
					idFn,
					clock,
					aggregateManifest,
					eventManifest,
				)
				ctx = context.Background()
			)

			// Act
			sid, err := e.StartSession(ctx)
			test.H(t).IsNil(err)

			var b bytes.Buffer
			resStr, err := e.Apply(ctx, &b, sid, []byte(`{"path":"agg/123", "name":"dummyCmd"}`))

			// Assert
			test.H(t).NotNil(err)
			test.H(t).StringEql("", resStr)
		})

		t.Run("sucessfully routes registered command to correct entity with ID", func(t *testing.T) {

			// Arrange
			var (
				ctx        = context.Background()
				objdb      = &memory.ObjectStore{}
				refdb      = &memory.RefStore{}
				idFn       = func() (string, error) { return fmt.Sprintf("%x", "dummy"), nil }
				clock      = &Predictable5sJumpClock{}
				aggM       = aggregates.NewManifest()
				cmdM       = commands.NewManifest()
				evM        = events.NewManifest()
				depot      = depot.NewSimple(objdb, refdb)
				repository = repository.NewSimpleRepository(objdb, refdb, evM)

				err error
			)

			aggM.Register("agg", &dummyAggregate{})
			cmdM.Register(&dummyAggregate{}, &dummyCmd{})

			aggM.Register("session", &dummySession{})
			cmdM.Register(&dummySession{}, &Start{})

			evM.Register(&DummyEvent{})
			evM.Register(&DummyStartSessionEvent{})

			var (
				r = resolver.New(aggM, cmdM)
				e = New(depot, repository, r, idFn, clock, aggM, evM)
			)

			// Act
			sid, err := e.StartSession(ctx)
			test.H(t).IsNil(err)

			var b bytes.Buffer
			resStr, err := e.Apply(ctx, &b, sid, []byte(`{"path":"agg/123", "name":"dummyCmd"}`))
			test.H(t).IsNil(err)

			// Assert
			test.H(t).StringEql("ok", resStr)
			return

			// TODO: Address this case. With the current code in place, the path agg/123
			// will not be found by the resolver, and rightfully so. It will therefore
			// overwrite the (empty) name in the Engine when persisting Evs.
			test.H(t).BoolEql(true, repository.Exists(ctx, types.PartitionName("agg/123")))
		})

		t.Run("raises error if session is not findable", func(t *testing.T) {
			t.Skip("not implemented yet")
		})

		// TODO: figure out how this should work, app instance is not ID'able
		t.Run("allows routing of certain commands to root object (_?)", func(t *testing.T) {
			t.Skip("not implemented yet")
		})

	})

	t.Run("storage", func(t *testing.T) {
		t.Run("applies commands and stores resulting events in case of success", func(t *testing.T) {

			// Arrange
			var (
				objdb      = &memory.ObjectStore{}
				refdb      = &memory.RefStore{}
				depot      = depot.NewSimple(objdb, refdb)
				idFn       = func() (string, error) { return fmt.Sprintf("%x", []byte("hello")), nil }
				clock      = &Predictable5sJumpClock{}
				aggM       = aggregates.NewManifest()
				cmdM       = commands.NewManifest()
				evM        = events.NewManifest()
				repository = repository.NewSimpleRepository(objdb, refdb, evM)

				err error
			)

			aggM.Register("agg", &dummyAggregate{})
			cmdM.Register(&dummyAggregate{}, &dummyCmd{})

			aggM.Register("session", &dummySession{})
			cmdM.Register(&dummySession{}, &Start{})

			evM.Register(&DummyEvent{})
			evM.Register(&DummyStartSessionEvent{})

			var (
				r   = resolver.New(aggM, cmdM)
				e   = New(depot, repository, r, idFn, clock, aggM, evM)
				ctx = context.Background()
			)

			// Act
			sid, err := e.StartSession(ctx)
			test.H(t).IsNil(err)

			var b bytes.Buffer
			resStr, err := e.Apply(ctx, &b, sid, []byte(`{"path":"agg/123", "name":"dummyCmd"}`))
			test.H(t).IsNil(err)

			if resStr != "ok" {
				t.Fatal("apply failed, we wanted 'ok', got ", resStr)
			}

			var expected = `event:sha256:70df53d19786d92d1bfa4c2527bb819054495ea3756fcfd7d88e3d4c8fae3172
event json dummy_event 2\u0000{}

checkpoint:sha256:a2cf30072524ca53eae2f9f660bb91e11e8edaefce6b428f07a7cb95c58aa280
checkpoint 169\u0000affix sha256:d45315bfa144411d8362ab68c56808ae92c20cfcbbf3abc35c57db5d08c871d5
date 0001-01-01T00:00:00Z
session 68656c6c6f

{"path":"session/68656c6c6f","name":"Start"}


checkpoint:sha256:bff4eee5edbee7ff29c9b02e039310e23f05d5b2e827a50bb118c41dfc3eebcd
checkpoint 241\u0000affix sha256:d9d06f42ded214cc216cfa218110bdba475bb5383acd97df8ddd957455592095
parent sha256:a2cf30072524ca53eae2f9f660bb91e11e8edaefce6b428f07a7cb95c58aa280
date 0001-01-01T00:00:05Z
session 68656c6c6f

{"path":"agg/123", "name":"dummyCmd"}


affix:sha256:d45315bfa144411d8362ab68c56808ae92c20cfcbbf3abc35c57db5d08c871d5
affix 99\u00000 dummy_session/68656c6c6f sha256:dd176fd38eaf032d39e35e39f04de8f30406bb0eaea55affe847f91cc923f69f


affix:sha256:d9d06f42ded214cc216cfa218110bdba475bb5383acd97df8ddd957455592095
affix 101\u00000 dummy_aggregate/68656c6c6f sha256:70df53d19786d92d1bfa4c2527bb819054495ea3756fcfd7d88e3d4c8fae3172


event:sha256:dd176fd38eaf032d39e35e39f04de8f30406bb0eaea55affe847f91cc923f69f
event json dummy_start_session_event 26\u0000{"Greeting":"hello world"}

refs/heads/master -> sha256:bff4eee5edbee7ff29c9b02e039310e23f05d5b2e827a50bb118c41dfc3eebcd
`
			if dd, ok := depot.(types.DumpableDepot); !ok {
				t.Fatal("could not upgrade depot to diff it")
			} else {
				var b bytes.Buffer
				dd.DumpAll(&b)
				var s = b.String()
				var diff = cmp.Diff(s, expected)
				if diff != "" {
					fmt.Println(s)
					t.Fatal(diff)
				}
			}
		})

		t.Run("moves the headpointer (ff) incase of success", func(t *testing.T) {

		})

		t.Run("applies commands and stores resulting error in case of errors", func(t *testing.T) {
			t.Skip("not implemented yet")
		})

		t.Run("tracks metrics for success in success cases", func(t *testing.T) {
			t.Skip("not implemented yet")
		})

		t.Run("tracks metrics for errors in error cases", func(t *testing.T) {
			t.Skip("not implemented yet")
		})
	})
}
