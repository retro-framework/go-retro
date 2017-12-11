package resolver

import (
	"context"
	"fmt"
	"testing"

	"github.com/leehambley/ls-cms/aggregates"
	"github.com/leehambley/ls-cms/commands"
	memory "github.com/leehambley/ls-cms/framework/in-memory"
	"github.com/leehambley/ls-cms/framework/types"
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

	emd := memory.NewEmptyDepot()

	aggm := aggregates.NewManifest()
	cmdm := commands.NewManifest()

	aggm.Register("agg", dummyAggregate{})
	cmdm.Register(dummyAggregate{}, &dummyCmd{})
	cmdm.Register(dummyAggregate{}, &otherDummyCmd{})

	r := resolver{aggm: aggm, cmdm: cmdm}

	r.Resolve(emd, []byte(`{"path":"agg", "name":"dummyCmd"}`))

	fmt.Println(r)

}

func Test_Resolver_ResolveExistingCmdRouteToNewInstance(t *testing.T) {
	t.Skip("for example sending to /session/ should gen a new session with id and treat that as the instance")
}
