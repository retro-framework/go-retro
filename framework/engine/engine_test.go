package engine

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/retro-framework/go-retro/aggregates"
	"github.com/retro-framework/go-retro/commands"
	memory "github.com/retro-framework/go-retro/framework/in-memory"
	"github.com/retro-framework/go-retro/framework/resolver"
	test "github.com/retro-framework/go-retro/framework/test_helper"
	"github.com/retro-framework/go-retro/framework/types"
)

type DummyEvent struct{}

type DummyStartSessionEvent struct {
	Greeting string
}

type dummySession struct {
}

func (_ *dummySession) ReactTo(types.Event) error {
	return nil
}

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

func (fssc *Start) Apply(context.Context, types.Aggregate, types.Depot) ([]types.Event, error) {
	return []types.Event{DummyStartSessionEvent{"hello world"}}, nil
}

type dummyAggregate struct {
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

func (dc *dummyCmd) Apply(_ context.Context, session types.Aggregate, _ types.Depot) ([]types.Event, error) {
	// if len(dc.s.seenEvents) != 2 {
	// 	return nil, errors.New(fmt.Sprintf("can't apply DummyEvent to dummyAggregate unless it has seen precisely two events so far (has seen %d)", len(dc.s.seenEvents)))
	// }
	dc.wasApplied = true
	return []types.Event{DummyEvent{}}, nil
}

// Sessions are a special case of aggregate We always need one, even if anon to
// do anything.
//
// Starting a session is a special-case of sending a command without a
// pre-existing session to the session aggregate to summon it into existence
func Test_Engine_StartSession(t *testing.T) {

	t.Run("creates a new session with parameters not matching an aggregate in the repository", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var (
			emd  = memory.NewEmptyDepot()
			idFn = func() (string, error) {
				return "123-stub-id", nil
			}
			resolveFn = func(ctx context.Context, depot types.Depot, cmd []byte) (types.CommandFunc, error) {
				fssc := Start{}
				fssc.SetState(&dummySession{})
				return fssc.Apply, nil
			}
			e = New(emd, resolveFn, idFn, nil)
		)

		// Act
		sid, err := e.StartSession(context.Background())
		test.H(t).BoolEql(true, emd.Exists(fmt.Sprintf("session/%s", sid)))
		test.H(t).IsNil(err)

		_, err = e.StartSession(context.Background())
		test.H(t).NotNil(err)
	})

	t.Run("persists the resulting session aggregate to the repository if the start command yields events", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var (
			emd  = memory.NewEmptyDepot()
			idFn = func() (string, error) {
				return fmt.Sprintf("%x", rand.Uint64()), nil
			}
			resolveFn = func(ctx context.Context, depot types.Depot, cmd []byte) (types.CommandFunc, error) {
				fssc := Start{}
				fssc.SetState(&dummySession{})
				return fssc.Apply, nil
			}
			e = New(emd, resolveFn, idFn, nil)
		)

		// Act
		sid, err := e.StartSession(context.Background())

		// Assert
		test.H(t).IsNil(err)
		test.H(t).BoolEql(true, emd.Exists(fmt.Sprintf("session/%s", sid)))
	})

	t.Run("forwards errors from the resolvefn to the caller", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var (
			resolverErr = fmt.Errorf("error from resolveFn")
			emd         = memory.NewEmptyDepot()
			idFn        = func() (string, error) {
				return fmt.Sprintf("%x", rand.Uint64()), nil
			}
			resolveFn = func(ctx context.Context, depot types.Depot, cmd []byte) (types.CommandFunc, error) {
				return nil, resolverErr
			}
		)
		var e = New(emd, resolveFn, idFn, nil)

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

	t.Run("channel", func(t *testing.T) {
		// setting something on context to get the "thread", should default to "main"
		// hint: git branching, should be an interface upgrade on the Depot like
		// "BranchableDepot" or something, I think.
		t.Skip("not implemented yet")
	})

	t.Run("routing", func(t *testing.T) {

		t.Run("raises error and logs it on unroutable (unregistered) command", func(t *testing.T) {
			t.Parallel()
			// Arrange
			var (
				emd = memory.NewEmptyDepot()

				idFn = func() (string, error) {
					return fmt.Sprintf("%x", rand.Uint64()), nil
				}

				aggm = aggregates.NewManifest()
				cmdm = commands.NewManifest()

				err error
			)

			// NOTE: no calls to register anything on the manifests except
			// the session!
			aggm.Register("session", &dummySession{})
			cmdm.Register(&dummySession{}, &Start{})

			var (
				r   = resolver.New(aggm, cmdm)
				e   = New(emd, r.Resolve, idFn, aggm)
				ctx = context.Background()
			)

			// Act
			sid, err := e.StartSession(ctx)
			test.H(t).IsNil(err)

			resStr, err := e.Apply(ctx, sid, []byte(`{"path":"agg/123", "name":"dummyCmd"}`))

			// Assert
			test.H(t).NotNil(err)
			test.H(t).StringEql("", resStr)
		})

		t.Run("sucessfully routes registered command to correct entity (with ID)", func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				emd = memory.NewEmptyDepot()

				idFn = func() (string, error) {
					return fmt.Sprintf("%x", rand.Uint64()), nil
				}

				aggm = aggregates.NewManifest()
				cmdm = commands.NewManifest()

				err error
			)

			aggm.Register("agg", &dummyAggregate{})
			cmdm.Register(&dummyAggregate{}, &dummyCmd{})

			aggm.Register("session", &dummySession{})
			cmdm.Register(&dummySession{}, &Start{})

			var (
				r   = resolver.New(aggm, cmdm)
				e   = New(emd, r.Resolve, idFn, aggm)
				ctx = context.Background()
			)

			// Act
			sid, err := e.StartSession(ctx)
			test.H(t).IsNil(err)

			resStr, err := e.Apply(ctx, sid, []byte(`{"path":"agg/123", "name":"dummyCmd"}`))
			test.H(t).IsNil(err)

			// Assert
			test.H(t).StringEql("ok", resStr)
			test.H(t).BoolEql(true, emd.Exists("agg/123"))
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
			t.Parallel()
			t.Skip("not implemented yet")
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
