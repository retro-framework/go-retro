package resolver

import (
	"context"
	"log"
	"testing"

	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
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

func Test_Resolver_ResolveExistingCmdSuccessfully(t *testing.T) {

	collector, err := zipkin.NewHTTPCollector("http://localhost:9411/api/v1/spans")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer collector.Close()

	tracer, err := zipkin.NewTracer(
		zipkin.NewRecorder(collector, true, "0.0.0.0:0", "example"),
	)
	if err != nil {
		log.Fatal(err)
	}
	opentracing.SetGlobalTracer(tracer)

	span, ctx := opentracing.StartSpanFromContext(context.Background(), "Test_Resolver_ResolveExistingCmdSuccessfully")
	defer span.Finish()

	// Arrange
	var (
		emd = memory.NewEmptyDepot()

		aggm = aggregates.NewManifest()
		cmdm = commands.NewManifest()

		dCmd = &dummyCmd{}
	)

	aggm.Register("agg", dummyAggregate{})
	cmdm.Register(dummyAggregate{}, dCmd)
	cmdm.Register(dummyAggregate{}, &otherDummyCmd{})

	r := resolver{aggm: aggm, cmdm: cmdm}

	// Act
	var (
		res types.CommandFunc
	)

	_ = res // TODO: Why do I need this line to avoid "is not used" error on the var decl above?
	res, err = r.Resolve(ctx, emd, []byte(`{"path":"agg", "name":"dummyCmd"}`))

	// Assert
	test.H(t).IsNil(err)
}

func Test_Resolver_ResolveExistingCmdRouteToNewInstance(t *testing.T) {
	t.Skip("for example sending to /session/ should gen a new session with id and treat that as the instance")
}
