package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/retro-framework/go-retro/aggregates"
	test "github.com/retro-framework/go-retro/framework/test_helper"
	"github.com/retro-framework/go-retro/framework/types"
)

type dummyCmd struct{}
type otherDummyCmd struct{ dummyCmd }

func (_ *dummyCmd) SetState(types.Aggregate) error { return nil }

func (_ *dummyCmd) Apply(context.Context, io.Writer, types.Session, types.Repository) (types.CommandResult, error) {
	return nil, nil
}

type dummyAggregate struct {
	aggregates.NamedAggregate
}

func (_ dummyAggregate) ReactTo(types.Event) error { return nil }

func Test_Commands_Register_TwiceSameCmdRaisesError(t *testing.T) {
	assertErrEql := test.H(t).ErrEql
	err := Register(&dummyAggregate{}, &dummyCmd{})
	assertErrEql(err, nil)
	err = Register(&dummyAggregate{}, &dummyCmd{})
	assertErrEql(err, fmt.Errorf("can't register command *commands.dummyCmd for aggregate commands.dummyAggregate, command already registered"))
}

type dummyArgs struct {
	Sentinel bool
	Name     string
}

func Test_Commands_RegisterWithArgs(t *testing.T) {
	var (
		m = DefaultManifest
	)
	dc := &otherDummyCmd{}

	err := m.RegisterWithArgs(&dummyAggregate{}, dc, dummyArgs{})
	if err != nil {
		t.Fatal("expected registering with args to work fine, got", err)
	}
	paramT, found := m.ArgTypeFor(dc)
	if !found {
		t.Fatal("expected to find arg type for", dc)
	}
	err = json.Unmarshal([]byte(`{"sentinel":true, "name":"test"}`), &paramT)
	if err != nil {
		t.Fatal("no error unmarshalling JSON got", err)
	}
	if da, ok := paramT.(*dummyArgs); !ok {
		t.Fatal("expected to be able to type assert paramT as dummyArgs")
	} else {
		if da.Sentinel != true {
			t.Fatal("expected paramt.sentinel to contain bool true")
		}
		if da.Name != "test" {
			t.Fatal("expected paramt.name to contain string test")
		}
	}
}
