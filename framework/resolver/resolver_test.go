package resolver

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/retro-framework/go-retro/aggregates"
	"github.com/retro-framework/go-retro/commands"
	"github.com/retro-framework/go-retro/framework/in-memory"
	test "github.com/retro-framework/go-retro/framework/test_helper"
	"github.com/retro-framework/go-retro/framework/types"
)

type OneEvent struct{}
type OtherEvent struct{}
type ExtraEvent struct{}

type dummyAggregate struct {
	seenEvents []types.Event
}

func (da *dummyAggregate) ReactTo(ev types.Event) error {
	da.seenEvents = append(da.seenEvents, ev)
	return nil
}

type dummySession struct{}

func (_ *dummySession) ReactTo(types.Event) error {
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

func (dc *dummyCmd) Apply(context.Context, types.Aggregate, types.Depot) ([]types.Event, error) {
	if len(dc.s.seenEvents) != 2 {
		return nil, errors.New(fmt.Sprintf("can't apply ExtraEvent to dummyAggregate unless it has seen precisely two events so far (has seen %d)", len(dc.s.seenEvents)))
	}
	dc.wasApplied = true
	return []types.Event{ExtraEvent{}}, nil
}

type otherDummyCmd struct {
	dummyCmd
}

type dummyCmdWithArgs struct {
	dummyCmd

	str string
	itn int
}

func (dcwa *dummyCmdWithArgs) SetArgs(args types.CommandArgs) error {

	// NOTE:
	//       JSON permits only float64 numbers in the JSON spec which
	//       places a burden on all implementors of CommandWithArgs.

	if strArg, ok := args["str"].(string); !ok {
		return errors.New("expected args[str] to be castable to string, it wasn't")
	} else {
		dcwa.str = strArg
	}

	if intArg, ok := args["int"].(float64); !ok {
		return errors.New("expected args[int] to be castable to float64, it wasn't")
	} else {
		dcwa.itn = int(intArg)
	}
	return nil
}

func Test_Resolver_AggregateLookup(t *testing.T) {

	t.Run("does not command to aggregate without ID", func(t *testing.T) {
		t.Parallel()
		// Arrange
		var (
			emd = memory.NewEmptyDepot()

			aggm = aggregates.NewManifest()
			cmdm = commands.NewManifest()

			err error
		)

		aggm.Register("agg", &dummyAggregate{})
		cmdm.Register(&dummyAggregate{}, &dummyCmd{})

		var r = resolver{aggm: aggm, cmdm: cmdm}

		// Act
		_, err = r.Resolve(context.Background(), emd, []byte(`{"path":"agg", "name":"dummyCmd"}`))

		// Assert
		test.H(t).NotNil(err)
		if rErr, ok := err.(Error); !ok {
			t.Fatal("could not cast err to Error")
		} else {
			test.H(t).StringEql("parse-agg-path", rErr.Op)
			test.H(t).StringEql("agg path \"agg\" does not split into exactly two parts", rErr.Err.Error())
		}
	})

	t.Run("resolves to an existing aggregate and retrieves it's history successfully", func(t *testing.T) {
		t.Parallel()
		// Arrange
		var (
			md = memory.NewDepot(map[string][]types.Event{"agg/123": []types.Event{OneEvent{}, OtherEvent{}}})

			aggm = aggregates.NewManifest()
			cmdm = commands.NewManifest()

			err error
		)

		aggm.Register("agg", &dummyAggregate{})
		cmdm.Register(&dummyAggregate{}, &dummyCmd{})
		cmdm.Register(&dummyAggregate{}, &otherDummyCmd{})

		var (
			r   = resolver{aggm: aggm, cmdm: cmdm}
			res types.CommandFunc
		)

		// Act
		res, err = r.Resolve(context.Background(), md, []byte(`{"path":"agg/123", "name":"dummyCmd"}`))

		// Assert
		test.H(t).IsNil(err)

		// Act
		newEvs, err := res(context.Background(), &dummySession{}, md)

		// Assert
		test.H(t).IsNil(err)
		test.H(t).IntEql(1, len(newEvs))
	})

	t.Run("returns empty aggregate in case of non-exixtent ID", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var (
			md = memory.NewDepot(map[string][]types.Event{"agg/456": []types.Event{OneEvent{}, OtherEvent{}}})
			//                                                 ^^^

			aggm = aggregates.NewManifest()
			cmdm = commands.NewManifest()

			err error
		)

		aggm.Register("agg", &dummyAggregate{})
		cmdm.Register(&dummyAggregate{}, &dummyCmd{})
		cmdm.Register(&dummyAggregate{}, &otherDummyCmd{})

		var (
			r   = resolver{aggm: aggm, cmdm: cmdm}
			res types.CommandFunc
		)

		// Act
		res, err = r.Resolve(context.Background(), md, []byte(`{"path":"agg/123", "name":"dummyCmd"}`))
		//                                                                  ^^^

		// Assert
		test.H(t).IsNil(err)

		// Act
		newEvs, err := res(context.Background(), &dummySession{}, md)

		// Assert
		test.H(t).NotNil(err) // dummyCmd throws error in case the aggregate has not !!= 2 events in the history
		test.H(t).IntEql(0, len(newEvs))
	})

}

func Test_Resolver_CommandParsing(t *testing.T) {

	t.Run("should raise an error if args are given and the command doesn't implement tyeps.CommandWithArgs", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var (
			md = memory.NewDepot(map[string][]types.Event{"agg/123": []types.Event{OneEvent{}, OtherEvent{}}})

			aggm = aggregates.NewManifest()
			cmdm = commands.NewManifest()

			err error
		)

		aggm.Register("agg", &dummyAggregate{})
		cmdm.Register(&dummyAggregate{}, &dummyCmd{})

		var (
			r = resolver{aggm: aggm, cmdm: cmdm}
		)

		// Act
		_, err = r.Resolve(context.Background(), md, []byte(`{"path":"agg/123", "name":"dummyCmd", "args":{"str": "bar", "int": 123}}`))

		// Assert
		test.H(t).NotNil(err) // dummyCmd does not implement the CommandWithArgs interface and "args" is given, this should
		test.H(t).StringEql(err.Error(), `resolver: op: "cast-cmd-with-args" err: "args given, but command does not implement CommandWithArgs"`)
	})

	t.Run("should parse cmd with args and set them on the object", func(t *testing.T) {

		t.Parallel()

		// Arrange
		var (
			md = memory.NewDepot(map[string][]types.Event{"agg/123": []types.Event{OneEvent{}, OtherEvent{}}})

			aggm = aggregates.NewManifest()
			cmdm = commands.NewManifest()

			dCmd = &dummyCmd{}

			err error
		)

		aggm.Register("agg", &dummyAggregate{})
		cmdm.Register(&dummyAggregate{}, dCmd)
		cmdm.Register(&dummyAggregate{}, &otherDummyCmd{})
		cmdm.Register(&dummyAggregate{}, &dummyCmdWithArgs{})

		var (
			r = resolver{aggm: aggm, cmdm: cmdm}
		)

		// Act
		_, err = r.Resolve(context.Background(), md, []byte(`{"path":"agg/123", "name":"dummyCmdWithArgs", "args":{"str": "bar", "int": 123}}`))

		// Assert
		test.H(t).IsNil(err)

	})

}

func Benchmark_Resolver_ResolveExistingCmdSuccessfully(b *testing.B) {

	b.ReportAllocs()

	// Arrange
	var (
		md = memory.NewDepot(map[string][]types.Event{"agg/123": []types.Event{OneEvent{}, OtherEvent{}}})

		aggm = aggregates.NewManifest()
		cmdm = commands.NewManifest()

		dCmd = &dummyCmd{}
	)

	aggm.Register("agg", &dummyAggregate{})
	cmdm.Register(&dummyAggregate{}, dCmd)
	cmdm.Register(&dummyAggregate{}, &otherDummyCmd{})

	var (
		r = resolver{aggm: aggm, cmdm: cmdm}
	)

	for n := 0; n < b.N; n++ {
		// Act
		r.Resolve(context.Background(), md, []byte(`{"path":"agg/123", "name":"dummyCmd"}`))
	}
}
