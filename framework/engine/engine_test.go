package engine

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"

	memory "github.com/retro-framework/go-retro/framework/in-memory"
	test "github.com/retro-framework/go-retro/framework/test_helper"
	"github.com/retro-framework/go-retro/framework/types"
)

type DummyStartSessionEvent struct {
	Greeting string
}

type dummySession struct {
}

func (_ *dummySession) ReactTo(types.Event) error {
	return nil
}

type fakeSessionStartCmd struct {
	s *dummySession
}

func (fssc *fakeSessionStartCmd) SetState(s types.Aggregate) error {
	if agg, ok := s.(*dummySession); ok {
		fssc.s = agg
		return nil
	} else {
		return errors.New("can't cast aggregate state")
	}
}

func (fssc *fakeSessionStartCmd) Apply(context.Context, types.Aggregate, types.Depot) ([]types.Event, error) {
	return []types.Event{DummyStartSessionEvent{"hello world"}}, nil
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
				fssc := fakeSessionStartCmd{}
				fssc.SetState(&dummySession{})
				return fssc.Apply, nil
			}
			e = NewEngine(emd, resolveFn, idFn)
		)

		// Act
		sid, err := e.StartSession(context.Background())
		test.H(t).BoolEql(true, emd.Exists(fmt.Sprintf("session/%s", sid)))
		test.H(t).IsNil(err)

		_, err = e.StartSession(context.Background())
		test.H(t).BoolEql(true, emd.Exists(fmt.Sprintf("session/%s", sid)))
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
				fssc := fakeSessionStartCmd{}
				fssc.SetState(&dummySession{})
				return fssc.Apply, nil
			}
			e = NewEngine(emd, resolveFn, idFn)
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
		var e = NewEngine(emd, resolveFn, idFn)

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
	t.Skip("not tested for yet")
}

func Test_Engine_Apply(t *testing.T) {

	t.Run("channel", func(t *testing.T) {
		// setting something on context to get the "thread", should default to "main"
		t.Skip("not implemented yet")
	})

	t.Run("routing", func(t *testing.T) {

		t.Run("raises error and logs it on unroutable (unregistered) command", func(t *testing.T) {
			t.Parallel()
		})

		t.Run("sucessfully routes registered command to correct entity (with ID)", func(t *testing.T) {
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
