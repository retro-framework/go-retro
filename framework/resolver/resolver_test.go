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
	"github.com/retro-framework/go-retro/framework/retro"
)

type OneEvent struct{}
type OtherEvent struct{}
type ExtraEvent struct{}

type dummyAggregate struct {
	aggregates.NamedAggregate
	seenEvents []retro.Event
}

func (da *dummyAggregate) ReactTo(ev retro.Event) error {
	da.seenEvents = append(da.seenEvents, ev)
	return nil
}

type dummySession struct {
	aggregates.NamedAggregate
}

func (_ *dummySession) ReactTo(retro.Event) error { return nil }

type dummyCmd struct {
	s          *dummyAggregate
	wasApplied bool
}

func (dc *dummyCmd) SetState(s retro.Aggregate) error {
	if agg, ok := s.(*dummyAggregate); ok {
		dc.s = agg
		return nil
	} else {
		return errors.New("can't cast aggregate state")
	}
}

func (dc *dummyCmd) Apply(context.Context, io.Writer, retro.Session, retro.Repo) (retro.CommandResult, error) {
	if len(dc.s.seenEvents) != 2 {
		return nil, errors.New(fmt.Sprintf("can't apply ExtraEvent to dummyAggregate unless it has seen precisely two events so far (has seen %d)", len(dc.s.seenEvents)))
	}
	dc.wasApplied = true
	return retro.CommandResult{dc.s: []retro.Event{ExtraEvent{}}}, nil
}

type dummyCmdWithArgs struct {
	dummyCmd

	args dummyCmdArgs
}

type dummyCmdArgs struct {
	str string
	i   int32 `json:"int"`
}

func (dcwa *dummyCmdWithArgs) SetArgs(args retro.CommandArgs) error {
	if typedArgs, ok := args.(*dummyCmdArgs); ok {
		dcwa.args = *typedArgs
	} else {
		return fmt.Errorf("can't typecast args")
	}
	return nil
}

func Test_commandDesc(t *testing.T) {
	var cd = commandDesc{Name: "HelloWorld", Path: ""}
	test.H(t).BoolEql(true, cd.DoesTargetRootAggregate())

	cd = commandDesc{Name: "HelloWorld", Path: "_"}
	test.H(t).BoolEql(true, cd.DoesTargetRootAggregate())

	cd = commandDesc{Name: "HelloWorld", Path: "foo/"}
	test.H(t).BoolEql(false, cd.DoesTargetRootAggregate())

	var errs, ok = commandDesc{Name: ""}.HasErrors()
	test.H(t).BoolEql(false, ok)
	test.H(t).IntEql(1, len(errs))
	test.H(t).StringEql("command name may not be empty", errs[0].Error())

	errs, ok = commandDesc{Name: "HelloWorld", Path: "foo/bar/baz"}.HasErrors()
	test.H(t).BoolEql(false, ok)
	test.H(t).IntEql(1, len(errs))
	test.H(t).StringEql("aggregate paths may not contain more than one forwardslash", errs[0].Error())

	errs, ok = commandDesc{Name: "HelloWorld", Path: "foo/"}.HasErrors()
	test.H(t).BoolEql(false, ok)
	test.H(t).IntEql(1, len(errs))
	test.H(t).StringEql("aggregate path must include an aggregate id after the first forwardslash", errs[0].Error())

}

func Test_Resolver_AggregateLookup(t *testing.T) {

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
				retro.EventFixture{
					agg: []retro.Event{
						OneEvent{},
						OtherEvent{},
					},
				},
			)
			r   = New(aggM, cmdM)
			res retro.Command

			err error
		)

		// Act pt. 1
		res, err = r.Resolve(ctx, repository, []byte(`{"path":"agg/123", "name":"dummyCmd"}`))

		// Assert pt. 1
		test.H(t).IsNil(err)

		// Act pt. 2
		var b bytes.Buffer
		newEvs, err := res.Apply(ctx, &b, &dummySession{}, repository)

		// Assert pt. 2
		test.H(t).IsNil(err)
		test.H(t).IntEql(1, len(newEvs))
	})

	t.Run("returns error aggregate in case of non-existent ID", func(t *testing.T) {

		t.Skip("not sure this is right")

		// Arrange
		var (
			aggM = aggregates.NewManifest()
			cmdM = commands.NewManifest()
		)
		aggM.Register("agg", &dummyAggregate{})
		cmdM.Register(&dummyAggregate{}, &dummyCmd{})

		var (
			ctx        = context.Background()
			repository = repository.NewEmptyMemory()
			r          = New(aggM, cmdM)

			err error
		)

		// Act
		_, err = r.Resolve(ctx, repository, []byte(`{"path":"agg/123", "name":"dummy_cmd"}`))
		//                                                       ^^^

		// Assert
		test.H(t).NotNil(err)
		test.H(t).StringEql(err.Error(), `resolver: op: "find-existing-aggregate" err: "no existing aggregate with name: agg/123"`)
	})

}

func Test_Resolver_CommandParsing(t *testing.T) {

	t.Run("should raise an error if args are given and the command doesn't implement tyeps.CommandWithArgs", func(t *testing.T) {

		// Arrange
		var (
			aggM = aggregates.NewManifest()
			cmdM = commands.NewManifest()
			agg  = &dummyAggregate{}
		)
		agg.SetName("agg/123")
		aggM.Register("agg", &dummyAggregate{})
		cmdM.Register(&dummyAggregate{}, &dummyCmd{})

		var (
			ctx        = context.Background()
			repository = repository.NewSimpleRepositoryDouble(retro.EventFixture{
				agg: []retro.Event{&OneEvent{}},
			})
			r   = New(aggM, cmdM)
			err error
		)

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
				retro.EventFixture{
					agg: []retro.Event{
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
