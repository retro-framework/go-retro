package test_helper

import (
	"testing"

	memory "github.com/leehambley/ls-cms/framework/in-memory"
	"github.com/leehambley/ls-cms/framework/types"
)

func StateFixture(t *testing.T, state map[string][]types.Event) types.Depot {
	t.Helper()
	return memory.NewDepot(state)
}

func EmptyStateFixture(t *testing.T) types.Depot {
	t.Helper()
	return memory.NewDepot(map[string][]types.Event{})
}

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

func (h helper) ErrEql(got, want error) {
	if got == nil && want == nil {
		return
	}
	if got != nil && want != nil {
		// todo mark as helper fn
		// if got.Error() != want.Error() {
		if got.Error() != want.Error() {
			h.t.Fatalf("boolean equality assertion failed, got %q wanted %q", got, want.Error())
		}
	}
}

func (h helper) NotNil(any interface{}) {
	if any == nil {
		h.t.Fatalf("wanted nil, got %v", any)
	}
}

func (h helper) BoolEql(got, want bool) {
	// todo mark as helper fn
	if got != want {
		h.t.Fatalf("boolean equality assertion failed, got %t wanted %t", got, want)
	}
}
