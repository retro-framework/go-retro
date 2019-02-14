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
	"github.com/retro-framework/go-retro/framework/resolver"
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

func (fssc *Start) Apply(context.Context, io.Writer, types.Session, types.Depot) (types.CommandResult, error) {
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

func (dc *dummyCmd) Apply(_ context.Context, _ io.Writer, _ types.Session, _ types.Depot) (types.CommandResult, error) {
	dc.wasApplied = true
	return types.CommandResult{dc.s: []types.Event{DummyEvent{}}}, nil
}

// Sessions are a special case of aggregate We always need one, even if anon to
// do anything.
//
// Starting a session is a special-case of sending a command without a
// pre-existing session to the session aggregate to summon it into existence
func Test_Engine_StartSession(t *testing.T) {

	t.Run("creates a new session with parameters not matching an aggregate in the repository", func(t *testing.T) {

		// Arrange
		var (
			ctx           = context.Background()
			depot         = depot.EmptySimpleMemory()
			eventManifest = events.NewManifest()
			idFn          = func() (string, error) { return "123-stub-id", nil }
			clock         = &Predictable5sJumpClock{}

			resolveFn = func(ctx context.Context, depot types.Depot, cmd []byte) (types.CommandFunc, error) {
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
		var e = New(depot, resolveFn, idFn, clock, aggregates.NewManifest(), eventManifest)

		// Act
		sid, err := e.StartSession(context.Background())

		// Assert
		test.H(t).IsNil(err)

		if dd, ok := depot.(types.DumpableDepot); !ok {
			t.Fatal("could not upgrade depot to diff it")
		} else {
			var b bytes.Buffer
			dd.DumpAll(&b)
		}

		test.H(t).BoolEql(true, depot.Exists(ctx, types.PartitionName(fmt.Sprintf("session/%s", sid))))
		_, err = e.StartSession(ctx)
		test.H(t).NotNil(err)
	})

	t.Run("persists the resulting session aggregate to the repository if the start command yields events", func(t *testing.T) {

		// Arrange
		var (
			ctx           = context.Background()
			depot         = depot.EmptySimpleMemory()
			eventManifest = events.NewManifest()
			idFn          = func() (string, error) { return "123-stub-id", nil }
			clock         = &Predictable5sJumpClock{}
			resolveFn     = func(ctx context.Context, depot types.Depot, cmd []byte) (types.CommandFunc, error) {
				var (
					fssc = Start{}
					s    = &dummySession{}
				)
				s.SetName(types.PartitionName("session/123-stub-id"))
				fssc.SetState(s)
				return fssc.Apply, nil
			}
		)
		eventManifest.Register(&DummyStartSessionEvent{})

		var e = New(depot, resolveFn, idFn, clock, aggregates.NewManifest(), eventManifest)

		// Act
		sid, err := e.StartSession(context.Background())

		// Assert
		test.H(t).IsNil(err)
		test.H(t).BoolEql(true, depot.Exists(ctx, types.PartitionName(fmt.Sprintf("session/%s", sid))))
	})

	t.Run("forwards errors from the resolvefn to the caller", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var (
			resolverErr = fmt.Errorf("error from resolveFn")
			depot       = depot.EmptySimpleMemory()
			idFn        = func() (string, error) { return fmt.Sprintf("%x", rand.Uint64()), nil }
			clock       = &Predictable5sJumpClock{}
			resolveFn   = func(ctx context.Context, depot types.Depot, cmd []byte) (types.CommandFunc, error) {
				return nil, resolverErr
			}
		)
		var e = New(depot, resolveFn, idFn, clock, aggregates.NewManifest(), events.NewManifest())

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

	t.Run("routing", func(t *testing.T) {

		t.Run("raises error and logs it on unroutable (unregistered) command", func(t *testing.T) {
			t.Parallel()
			// Arrange

			var manifests = struct {
				event     types.EventManifest
				aggregate types.AggregateManifest
				command   types.CommandManifest
			}{
				aggregate: aggregates.NewManifest(),
				command:   commands.NewManifest(),
				event:     events.NewManifest(),
			}

			var (
				depot = depot.EmptySimpleMemory()
				idFn  = func() (string, error) {
					return fmt.Sprintf("%x", rand.Uint64()), nil
				}
				clock = &Predictable5sJumpClock{}
				err   error
			)

			// NOTE: no calls to register anything on the manifests except
			// the session!
			manifests.aggregate.Register("session", &dummySession{})
			manifests.command.Register(&dummySession{}, &Start{})
			manifests.event.Register(&DummyStartSessionEvent{})

			var (
				r = resolver.New(manifests.aggregate, manifests.command)
				e = New(
					depot,
					r.Resolve,
					idFn,
					clock,
					manifests.aggregate,
					manifests.event,
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
			t.Parallel()

			// Arrange
			var (
				ctx   = context.Background()
				depot = depot.EmptySimpleMemory()
				idFn  = func() (string, error) {
					return fmt.Sprintf("%x", rand.Uint64()), nil
				}
				clock = &Predictable5sJumpClock{}

				aggm = aggregates.NewManifest()
				cmdm = commands.NewManifest()
				evm  = events.NewManifest()

				err error
			)

			aggm.Register("agg", &dummyAggregate{})
			cmdm.Register(&dummyAggregate{}, &dummyCmd{})

			aggm.Register("session", &dummySession{})
			cmdm.Register(&dummySession{}, &Start{})

			evm.Register(&DummyEvent{})
			evm.Register(&DummyStartSessionEvent{})

			var (
				r = resolver.New(aggm, cmdm)
				e = New(depot, r.Resolve, idFn, clock, aggm, evm)
			)

			// Act
			sid, err := e.StartSession(ctx)
			test.H(t).IsNil(err)

			var b bytes.Buffer
			resStr, err := e.Apply(ctx, &b, sid, []byte(`{"path":"agg/123", "name":"dummyCmd"}`))
			test.H(t).IsNil(err)

			// Assert
			test.H(t).StringEql("ok", resStr)
			test.H(t).BoolEql(true, depot.Exists(ctx, types.PartitionName("agg/123")))
		})

		t.Run("raises error if session is not findable", func(t *testing.T) {
			t.Parallel()
			t.Skip("not implemented yet")
		})

		// TODO: figure out how this should work, app instance is not ID'able
		t.Run("allows routing of certain commands to root object (_?)", func(t *testing.T) {
			t.Parallel()
			t.Skip("not implemented yet")
		})

	})

	t.Run("storage", func(t *testing.T) {
		t.Run("applies commands and stores resulting events in case of success", func(t *testing.T) {

			// Arrange
			var (
				depot = depot.EmptySimpleMemory()

				idFn = func() (string, error) {
					return fmt.Sprintf("%x", []byte("hello")), nil
					// return fmt.Sprintf("%x", rand.Uint64()), nil
				}

				clock = &Predictable5sJumpClock{}

				aggm = aggregates.NewManifest()
				cmdm = commands.NewManifest()
				evm  = events.NewManifest()

				err error
			)

			aggm.Register("agg", &dummyAggregate{})
			cmdm.Register(&dummyAggregate{}, &dummyCmd{})

			aggm.Register("session", &dummySession{})
			cmdm.Register(&dummySession{}, &Start{})

			evm.Register(&DummyEvent{})
			evm.Register(&DummyStartSessionEvent{})

			var (
				r   = resolver.New(aggm, cmdm)
				e   = New(depot, r.Resolve, idFn, clock, aggm, evm)
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


event:sha256:5447b1b56906b752fa19be063493ad8651d0ec6bff0aaf54c9fb7ea1cc2f19ca
event json DummyEvent 2\u0000{}

checkpoint:sha256:babdee29734925d4ffba9fbfccf02b54e86abdb1cdca2cd6ee1994a459066180
checkpoint 250\u0000affix sha256:2b1a392c0a69cbaf9db010fdb2642e504714e06cff387befdc509a016114cdb7
parent sha256:d9602dc0f0deed34457705e1d6b28386bd9946344c8f98ff7e73ad5dcb4fb449
date Mon, 01 Jan 0001 00:00:05 UTC
session 68656c6c6f

{"path":"agg/123", "name":"dummyCmd"}


checkpoint:sha256:d9602dc0f0deed34457705e1d6b28386bd9946344c8f98ff7e73ad5dcb4fb449
checkpoint 178\u0000affix sha256:eb7f7b20d4172478ba96ada60e00ff3675998bcac7b76768a7c25c012e64b310
date Mon, 01 Jan 0001 00:00:00 UTC
session 68656c6c6f

{"path":"session/68656c6c6f","name":"Start"}


affix:sha256:eb7f7b20d4172478ba96ada60e00ff3675998bcac7b76768a7c25c012e64b310
affix 93\u00000 session/68656c6c6f sha256:190dc13864f5a6622b97cf7faf77f8717588c5d95774cf4fb2a6a0a6e8923bb2


refs/heads/master -> sha256:babdee29734925d4ffba9fbfccf02b54e86abdb1cdca2cd6ee1994a459066180
`
			if dd, ok := depot.(types.DumpableDepot); !ok {
				t.Fatal("could not upgrade depot to diff it")
			} else {
				var b bytes.Buffer
				dd.DumpAll(&b)
				var s = b.String()
				var diff = cmp.Diff(s, expected)
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})

		t.Run("moves the headpointer (ff) incase of success", func(t *testing.T) {

		})

		t.Run("applies commands and stores resulting error in case of errors", func(t *testing.T) {
			t.Parallel()
			t.Skip("not implemented yet")
		})

		t.Run("tracks metrics for success in success cases", func(t *testing.T) {
			t.Parallel()
			t.Skip("not implemented yet")
		})

		t.Run("tracks metrics for errors in error cases", func(t *testing.T) {
			t.Parallel()
			t.Skip("not implemented yet")
		})
	})
}
