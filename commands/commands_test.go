package commands

import (
	"context"
	"fmt"
	"testing"

	test "github.com/leehambley/ls-cms/framework/test_helper"
	"github.com/leehambley/ls-cms/framework/types"
)

type dummyCmd struct{}
type otheDummyCmd struct{ dummyCmd }

func (_ dummyCmd) Apply(context.Context, types.Aggregate, types.Depot) ([]types.Event, error) {
	return nil, nil
}

type dummyAggregate struct{}

func (_ dummyAggregate) ReactTo(types.Event) error { return nil }

func Test_Commands_Register_TwiceSameCmdRaisesError(t *testing.T) {
	assertErrEql := test.H(t).ErrEql
	err := Register(dummyAggregate{}, dummyCmd{})
	assertErrEql(err, nil)
	err = Register(dummyAggregate{}, dummyCmd{})
	assertErrEql(err, fmt.Errorf("Can't register command commands.dummyCmd for aggregate commands.dummyAggregate, command already registered"))
}
