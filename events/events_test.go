package events

import (
	"testing"

	test "github.com/leehambley/ls-cms/framework/test_helper"
)

type dummyEv struct{}

func Test_Events_RegisterEv_TwiceSameEvRaisesError(t *testing.T) {
	assertNotNil := test.H(t).NotNil
	assertErrEql := test.H(t).ErrEql
	err := Register(&dummyEv{})
	assertErrEql(err, nil)
	err = Register(&dummyEv{})
	assertNotNil(err)
}

func Test_Events_RegisterEv_TakesEitherPointerOrConcrete(t *testing.T) {
	t.Skip("can register's use of reflect not handle concrete types yet?")
	// assertErrEql := test.H(t).ErrEql
	// _ = Register(dummyEv{})
	// assertErrEql(err, nil)
	// err = Register(&dummyEv{})
	// assertErrEql(err, fmt.Errorf("Can't register command commands.dummyCmd for aggregate commands.dummyAggregate, command already registered"))
}
