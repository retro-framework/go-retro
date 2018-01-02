package resolver

import (
	"context"
	"testing"

	"github.com/retro-framework/go-retro/aggregates"
	"github.com/retro-framework/go-retro/commands"
	"github.com/retro-framework/go-retro/framework/in-memory"
	test "github.com/retro-framework/go-retro/framework/test_helper"
	"github.com/retro-framework/go-retro/framework/types"
)

type dummyAggregate struct{}

func (_ dummyAggregate) ReactTo(types.Event) error {
	return nil
}

type dummyCmd struct {
	wasApplied bool
}
type otherDummyCmd struct{ dummyCmd }

func (dc *dummyCmd) Apply(context.Context, types.Aggregate, types.Depot) ([]types.Event, error) {
	dc.wasApplied = true
	return nil, nil
}

func Test_Resolver_DoesNotResolveCmdToAggregateWithoutID(t *testing.T) {
	var (
		emd = memory.NewEmptyDepot()

		aggm = aggregates.NewManifest()
		cmdm = commands.NewManifest()
		dCmd = &dummyCmd{}

		err error
	)

	aggm.Register("agg", dummyAggregate{})
	cmdm.Register(dummyAggregate{}, dCmd)

	r := resolver{aggm: aggm, cmdm: cmdm}

	_, err = r.Resolve(context.Background(), emd, []byte(`{"path":"agg", "name":"dummyCmd"}`))

	test.H(t).NotNil(err)

	if rErr, ok := err.(Error); !ok {
		t.Fatal("could not cast err to Error")
	} else {
		test.H(t).StringEql("parse-agg-path", rErr.Op)
		test.H(t).StringEql("agg path \"agg\" does not split into exactly two parts", rErr.Err.Error())
	}
}

func Test_Resolver_ResolveExistingCmdToExistingAggregateSuccessfully(t *testing.T) {

	// collector, err := zipkin.NewHTTPCollector("http://localhost:9411/api/v1/spans")
	// if err != nil {
	// 	log.Fatal(err)
	// 	return
	// }
	// defer collector.Close()

	// tracer, err := zipkin.NewTracer(
	// 	zipkin.NewRecorder(collector, true, "0.0.0.0:0", "example"),
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// opentracing.SetGlobalTracer(tracer)

	// span, ctx := opentracing.StartSpanFromContext(context.Background(), "Test_Resolver_ResolveExistingCmdSuccessfully")
	// defer span.Finish()

	// Arrange
	var (
		emd = memory.NewEmptyDepot()

		aggm = aggregates.NewManifest()
		cmdm = commands.NewManifest()

		dCmd = &dummyCmd{}

		err error
	)

	aggm.Register("agg", dummyAggregate{})
	cmdm.Register(dummyAggregate{}, dCmd)
	cmdm.Register(dummyAggregate{}, &otherDummyCmd{})

	r := resolver{aggm: aggm, cmdm: cmdm}

	// Act
	var res types.CommandFunc

	res, err = r.Resolve(context.Background(), emd, []byte(`{"path":"agg/123", "name":"dummyCmd"}`))

	// Assert
	test.H(t).IsNil(err)
	test.H(t).NotNil(res) // assignment to interface type is also a useful assertion
}

func Test_Resolver_ResolveExistingCmdToNonExistantAggregateSuccessfully(t *testing.T) {
	t.Skip("for example sending to /session/ should gen a new session with id and treat that as the instance")
}
