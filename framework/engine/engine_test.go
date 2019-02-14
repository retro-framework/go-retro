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

type dummySession struct {
	aggregates.NamedAggregate
}

func (_ *dummySession) ReactTo(types.Event) error { return nil }

type Start struct {
	s *dummySession
}

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
			resolveFn  = func(_ context.Context, _ types.Repository, _ []byte) (types.CommandFunc, error) {
				var (
					fssc = Start{}
					s    = &dummySession{}
				)
				s.SetName(types.PartitionName("session/123-stub-id"))
				fssc.SetState(s)
				return fssc.Apply, nil
			}
		)
		evM.Register(&DummyEvent{})
		evM.Register(&DummyStartSessionEvent{})
		var e = New(depot, repository, resolveFn, idFn, clock, aggregates.NewManifest(), evM)

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
			resolveFn  = func(_ context.Context, _ types.Repository, _ []byte) (types.CommandFunc, error) {
				var (
					fssc = Start{}
					s    = &dummySession{}
				)
				s.SetName(types.PartitionName("session/123-stub-id"))
				fssc.SetState(s)
				return fssc.Apply, nil
			}
		)
		evM.Register(&DummyEvent{})
		evM.Register(&DummyStartSessionEvent{})
		var e = New(depot, repository, resolveFn, idFn, clock, aggregates.NewManifest(), evM)

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
			resolveFn  = func(_ context.Context, _ types.Repository, _ []byte) (types.CommandFunc, error) {
				var (
					fssc = Start{}
					s    = &dummySession{}
				)
				s.SetName(types.PartitionName("session/123-stub-id"))
				fssc.SetState(s)
				return fssc.Apply, nil
			}
		)
		evM.Register(&DummyStartSessionEvent{})

		var e = New(depot, repository, resolveFn, idFn, clock, aggregates.NewManifest(), evM)

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
			resolveFn     = func(_ context.Context, _ types.Repository, _ []byte) (types.CommandFunc, error) {
				return nil, resolverErr
			}
		)
		var e = New(depot, repository, resolveFn, idFn, clock, aggregates.NewManifest(), events.NewManifest())

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

			resolveFn = func(_ context.Context, _ types.Repository, _ []byte) (types.CommandFunc, error) {
				var (
					fssc = Start{}
					s    = &dummySession{}
				)
				s.SetName(types.PartitionName("session/123-stub-id"))
				fssc.SetState(s)
				return fssc.Apply, nil
			}
		)
		eventManifest.Register(&DummyEvent{})
		eventManifest.Register(&DummyStartSessionEvent{})
		var e = New(depot, repository, resolveFn, idFn, clock, aggregates.NewManifest(), eventManifest)

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
				r = resolver.New(aggregateManifest, commandManifest)
				e = New(
					depot,
					repository,
					r.Resolve,
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

		t.Run("sucessfully routes registered command to correct entity (with ID)", func(t *testing.T) {

			// Arrange
			var (
				ctx        = context.Background()
				objdb      = &memory.ObjectStore{}
				refdb      = &memory.RefStore{}
				depot      = depot.NewSimple(objdb, refdb)
				idFn       = func() (string, error) { return fmt.Sprintf("%x", rand.Uint64()), nil }
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
				r = resolver.New(aggM, cmdM)
				e = New(depot, repository, r.Resolve, idFn, clock, aggM, evM)
			)

			// Act
			sid, err := e.StartSession(ctx)
			test.H(t).IsNil(err)

			var b bytes.Buffer
			resStr, err := e.Apply(ctx, &b, sid, []byte(`{"path":"agg/123", "name":"dummyCmd"}`))
			test.H(t).IsNil(err)

			// Assert
			test.H(t).StringEql("ok", resStr)
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
				e   = New(depot, repository, r.Resolve, idFn, clock, aggM, evM)
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

			var expected = `event:sha256:190dc13864f5a6622b97cf7faf77f8717588c5d95774cf4fb2a6a0a6e8923bb2
event json DummyStartSessionEvent 26\u0000{"Greeting":"hello world"}

affix:sha256:2b1a392c0a69cbaf9db010fdb2642e504714e06cff387befdc509a016114cdb7
affix 82\u00000 agg/123 sha256:5447b1b56906b752fa19be063493ad8651d0ec6bff0aaf54c9fb7ea1cc2f19ca


checkpoint:sha256:383bca8be8bf2fc851deb8c437678b27d96c236b28002953ad832d8f3433c674
checkpoint 169\u0000affix sha256:eb7f7b20d4172478ba96ada60e00ff3675998bcac7b76768a7c25c012e64b310
date 0001-01-01T00:00:00Z
session 68656c6c6f

{"path":"session/68656c6c6f","name":"Start"}


checkpoint:sha256:476cc2981a4e452ece7299c7c3bc3cfc11f722122c1cc6340e352ca597d8a4a8
checkpoint 241\u0000affix sha256:2b1a392c0a69cbaf9db010fdb2642e504714e06cff387befdc509a016114cdb7
parent sha256:383bca8be8bf2fc851deb8c437678b27d96c236b28002953ad832d8f3433c674
date 0001-01-01T00:00:05Z
session 68656c6c6f

{"path":"agg/123", "name":"dummyCmd"}


event:sha256:5447b1b56906b752fa19be063493ad8651d0ec6bff0aaf54c9fb7ea1cc2f19ca
event json DummyEvent 2\u0000{}

affix:sha256:eb7f7b20d4172478ba96ada60e00ff3675998bcac7b76768a7c25c012e64b310
affix 93\u00000 session/68656c6c6f sha256:190dc13864f5a6622b97cf7faf77f8717588c5d95774cf4fb2a6a0a6e8923bb2


refs/heads/master -> sha256:476cc2981a4e452ece7299c7c3bc3cfc11f722122c1cc6340e352ca597d8a4a8
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
