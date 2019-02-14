// +build unit

package resolver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/retro-framework/go-retro/aggregates"
	"github.com/retro-framework/go-retro/commands"
	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/repository"
	"github.com/retro-framework/go-retro/framework/storage/memory"
	test "github.com/retro-framework/go-retro/framework/test_helper"
	"github.com/retro-framework/go-retro/framework/types"
)

type OneEvent struct{}
type OtherEvent struct{}
type ExtraEvent struct{}

type dummyAggregate struct {
	aggregates.NamedAggregate
	seenEvents []types.Event
}

func (da *dummyAggregate) ReactTo(ev types.Event) error {
	da.seenEvents = append(da.seenEvents, ev)
	return nil
}

type dummySession struct {
	aggregates.NamedAggregate
}

func (_ *dummySession) ReactTo(types.Event) error { return nil }

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

func (dc *dummyCmd) Apply(context.Context, io.Writer, types.Session, types.Repository) (types.CommandResult, error) {
	if len(dc.s.seenEvents) != 2 {
		return nil, errors.New(fmt.Sprintf("can't apply ExtraEvent to dummyAggregate unless it has seen precisely two events so far (has seen %d)", len(dc.s.seenEvents)))
	}
	dc.wasApplied = true
	return types.CommandResult{dc.s: []types.Event{ExtraEvent{}}}, nil
}

type dummyCmdWithArgs struct {
	dummyCmd

	args dummyCmdArgs
}

type dummyCmdArgs struct {
	str string
	i   int32 `json:"int"`
}

func (dcwa *dummyCmdWithArgs) SetArgs(args types.CommandArgs) error {
	if typedArgs, ok := args.(*dummyCmdArgs); ok {
		dcwa.args = *typedArgs
	} else {
		return fmt.Errorf("can't typecast args")
	}
	return nil
}

func Test_Resolver_AggregateLookup(t *testing.T) {

	t.Parallel()

	t.Run("does not resolve command to aggregate without ID", func(t *testing.T) {

		// Arrange
		var (
			ctx        = context.Background()
			objdb      = &memory.ObjectStore{}
			refdb      = &memory.RefStore{}
			evM        = events.NewManifest()
			aggM       = aggregates.NewManifest()
			cmdM       = commands.NewManifest()
			repository = repository.NewSimpleRepository(objdb, refdb, evM)
			r          = New(aggM, cmdM)

			err error
		)
		aggM.Register("agg", &dummyAggregate{})
		cmdM.Register(&dummyAggregate{}, &dummyCmd{})

		// Act
		_, err = r.Resolve(ctx, repository, []byte(`{"path":"agg", "name":"dummyCmd"}`))

		// Assert
		test.H(t).NotNil(err)
		if rErr, ok := err.(Error); !ok {
			t.Fatal("could not cast err to Error")
		} else {
			test.H(t).StringEql("parse-agg-path", rErr.Op)
			test.H(t).StringEql("agg path \"agg\" does not split into exactly two parts", rErr.Err.Error())
		}
	})

	t.Run("resolves to an existing aggregate and retrieves its history successfully", func(t *testing.T) {

		// Arrange
		var (
			agg  = &dummyAggregate{}
			evM  = events.NewManifest()
			aggM = aggregates.NewManifest()
			cmdM = commands.NewManifest()
		)
		aggM.Register("agg", &dummyAggregate{})
		cmdM.Register(&dummyAggregate{}, &dummyCmd{})
		evM.Register(&OneEvent{})
		evM.Register(&OtherEvent{})

		agg.SetName("agg/123")

		var (
			ctx        = context.Background()
			repository = repository.NewSimpleRepositoryDouble(
				types.EventFixture{
					agg: []types.Event{
						OneEvent{},
						OtherEvent{},
					},
				},
			)
			r   = New(aggM, cmdM)
			res types.CommandFunc

			err error
		)

		// Act pt. 1
		res, err = r.Resolve(context.Background(), repository, []byte(`{"path":"agg/123", "name":"dummyCmd"}`))

		// Assert pt. 1
		test.H(t).IsNil(err)

		// Act pt. 2
		var b bytes.Buffer
		newEvs, err := res(ctx, &b, &dummySession{}, repository)

		// Assert pt. 2
		test.H(t).IsNil(err)
		test.H(t).IntEql(1, len(newEvs))
	})

	t.Run("returns empty aggregate in case of non-exixtent ID", func(t *testing.T) {

		// Arrange
		var (
			agg  = &dummyAggregate{}
			evM  = events.NewManifest()
			aggM = aggregates.NewManifest()
			cmdM = commands.NewManifest()
		)
		aggM.Register("agg", &dummyAggregate{})
		cmdM.Register(&dummyAggregate{}, &dummyCmd{})
		evM.Register(&OneEvent{})
		evM.Register(&OtherEvent{})

		agg.SetName("agg/456")
		//               ^^^ (!= 123 below)

		var (
			ctx        = context.Background()
			repository = repository.NewSimpleRepositoryDouble(
				types.EventFixture{
					agg: []types.Event{
						OneEvent{},
						OtherEvent{},
					},
				},
			)
			r   = New(aggM, cmdM)
			res types.CommandFunc

			err error
		)

		// Act pt. 1
		res, err = r.Resolve(context.Background(), repository, []byte(`{"path":"agg/123", "name":"dummyCmd"}`))
		//                                                                  ^^^

		// Assert pt. 1
		test.H(t).IsNil(err)

		// Act pt. 2
		var b bytes.Buffer
		newEvs, err := res(ctx, &b, &dummySession{}, repository)

		// Assert pt. 2
		test.H(t).NotNil(err) // dummyCmd throws error in case the aggregate has not !!= 2 events in the history
		test.H(t).IntEql(0, len(newEvs))
	})

}

func Test_Resolver_CommandParsing(t *testing.T) {

	t.Run("should raise an error if args are given and the command doesn't implement tyeps.CommandWithArgs", func(t *testing.T) {

		// Arrange
		var (
			ctx        = context.Background()
			objdb      = &memory.ObjectStore{}
			refdb      = &memory.RefStore{}
			aggM       = aggregates.NewManifest()
			cmdM       = commands.NewManifest()
			evM        = events.NewManifest()
			repository = repository.NewSimpleRepository(objdb, refdb, evM)
			r          = New(aggM, cmdM)

			err error
		)
		aggM.Register("agg", &dummyAggregate{})
		cmdM.Register(&dummyAggregate{}, &dummyCmd{})

		// Act
		_, err = r.Resolve(ctx, repository, []byte(`{"path":"agg/123", "name":"dummyCmd", "args":{"str": "bar", "int": 123}}`))

		// Assert
		test.H(t).NotNil(err) // dummyCmd does not implement the CommandWithArgs interface and "args" is given, this should
		test.H(t).StringEql(err.Error(), `resolver: op: "cast-cmd-with-args" err: "args given, but command does not implement CommandWithArgs"`)
	})

	t.Run("should parse cmd with args and set them on the object", func(t *testing.T) {

		// Arrange
		var (
			ctx  = context.Background()
			agg  = &dummyAggregate{}
			evM  = events.NewManifest()
			aggM = aggregates.NewManifest()
			cmdM = commands.NewManifest()
		)
		aggM.Register("agg", &dummyAggregate{})
		cmdM.Register(&dummyAggregate{}, &dummyCmd{})
		cmdM.RegisterWithArgs(&dummyAggregate{}, &dummyCmdWithArgs{}, &dummyCmdArgs{})
		evM.Register(&OneEvent{})
		evM.Register(&OtherEvent{})
		agg.SetName("agg/123")

		var (
			repository = repository.NewSimpleRepositoryDouble(
				types.EventFixture{
					agg: []types.Event{
						OneEvent{},
						OtherEvent{},
					},
				},
			)
			r = New(aggM, cmdM)

			err error
		)

		// Act
		_, err = r.Resolve(ctx, repository, []byte(`{"path":"agg/123", "name":"dummyCmdWithArgs", "args":{"str": "bar", "int": 123}}`))

		// Assert
		test.H(t).IsNil(err)

	})

}

func Benchmark_Resolver_ResolveExistingCmdSuccessfully(b *testing.B) {

	b.ReportAllocs()

	// Arrange
	var (
		ctx        = context.Background()
		aggM       = aggregates.NewManifest()
		cmdM       = commands.NewManifest()
		evM        = events.NewManifest()
		objdb      = &memory.ObjectStore{}
		refdb      = &memory.RefStore{}
		repository = repository.NewSimpleRepository(objdb, refdb, evM)
		r          = New(aggM, cmdM)
	)

	aggM.Register("agg", &dummyAggregate{})
	cmdM.Register(&dummyAggregate{}, &dummyCmd{})

	for n := 0; n < b.N; n++ {
		// Act
		r.Resolve(ctx, repository, []byte(`{"path":"agg/123", "name":"dummyCmd"}`))
	}
}
