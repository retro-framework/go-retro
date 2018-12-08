package test_helper

import (
	"fmt"
	"github.com/google/go-cmp/cmp" // "github.com/retro-framework/go-retro/framework/storage/memory"
	"github.com/retro-framework/go-retro/framework/types"
	"strings"
	"testing"
)

// func StateFixture(t *testing.T, state map[string][]types.Event) types.Depot {
// 	t.Helper()
// 	return memory.NewDepot(state)
// }

// func EmptyStateFixture(t *testing.T) types.Depot {
// 	t.Helper()
// 	return memory.NewDepot(map[string][]types.Event{})
// }

func AggStateFixture(name string, evs ...types.Event) map[string][]types.Event {
	// TODO ensure that name isn't empty
	return map[string][]types.Event{name: evs}
}

func H(t *testing.T) helper {
	t.Helper()
	return helper{t}
}

type helper struct {
	t *testing.T
}

func (h helper) TypeEql(got, want interface{}) {
	h.t.Helper()
	// check obvious case
	if got == nil && want == nil {
		return
	}
	// check for type equality
	if strings.Compare(fmt.Sprintf("%T", got), fmt.Sprintf("%T", want)) != 0 {
		h.t.Fatalf("type equality assertion failed, got %q wanted %q", fmt.Sprintf("%T", got), fmt.Sprintf("%T", want))
	}
}

func (h helper) IntEql(got, want int) {
	h.t.Helper()
	if got != want {
		h.t.Fatalf("int equality assertion failed, got %d wanted %d", got, want)
	}
}

func (h helper) StringEql(got, want string) {
	h.t.Helper()
	if diff := cmp.Diff(want, got); diff != "" {
		h.t.Errorf("string equality assertion failed (-got +want)\n%s", diff)
	}
}

func (h helper) InterfaceEql(got, want interface{}) {
	h.t.Helper()
	if diff := cmp.Diff(want, got); diff != "" {
		h.t.Errorf("string equality assertion failed (-got +want)\n%s", diff)
	}
}

func (h helper) ErrEql(got, want error) {
	h.t.Helper()
	if got == nil && want == nil {
		return
	}
	if got != nil && want != nil {
		if got.Error() != want.Error() {
			h.t.Fatalf("error equality assertion failed, got %q wanted %q", got, want.Error())
		}
	}
}

func (h helper) IsNil(any interface{}) {
	h.t.Helper()
	if any != nil {
		h.t.Fatalf("wanted not nil, got %v", any)
	}
}

func (h helper) NotNil(any interface{}) {
	h.t.Helper()
	if any == nil {
		h.t.Fatalf("wanted not nil, got %v", any)
	}
}

func (h helper) BoolEql(got, want bool) {
	h.t.Helper()
	if got != want {
		h.t.Fatalf("boolean equality assertion failed, got %t wanted %t", got, want)
	}
}
