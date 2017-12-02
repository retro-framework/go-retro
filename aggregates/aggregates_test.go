package aggregates

import (
	"testing"

	test "github.com/leehambley/ls-cms/framework/test_helper"
	"github.com/leehambley/ls-cms/framework/types"
)

type dummyAggregate struct{}

func (_ dummyAggregate) ReactTo(types.Event) error { return nil }

func Test_Aggregates_Register_TwiceSameEvRaisesError(t *testing.T) {
	assertNotNil := test.H(t).NotNil
	assertErrEql := test.H(t).ErrEql
	err := Register("_", &dummyAggregate{})
	assertErrEql(err, nil)
	err = Register("_", &dummyAggregate{})
	assertNotNil(err)
	// assertErrEql(err, fmt.Errorf("Can't register command commands.dummyCmd for aggregate commands.dummyAggregate, command already registered"))
}
