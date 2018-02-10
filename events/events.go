package events

import (
	"fmt"
	"reflect"

	"github.com/retro-framework/go-retro/framework/types"
)

var DefaultManifest = NewManifest()

type manifest struct {
	m map[string]reflect.Type
}

func NewManifest() types.EventManifest {
	return &manifest{make(map[string]reflect.Type)}
}

func Register(ev types.Event) error {
	return DefaultManifest.Register(ev)
}

func (m *manifest) RegisterAs(evName string, ev types.Event) error {
	var v = m.toType(ev)
	if _, exists := m.m[evName]; exists {
		return fmt.Errorf("can't register event %s, already registered (event names are not namespaced)", evName)
	}
	m.m[evName] = v
	return nil
}

func (m *manifest) Register(ev types.Event) error {
	var (
		v      = m.toType(ev)
		evName = v.Name()
	)
	return m.RegisterAs(evName, ev)
}

// TODO: should return an error if we can't map the ev type
func (m *manifest) KeyFor(ev types.Event) string {
	for name, tepy := range m.m {
		if tepy == m.toType(ev) {
			return name
		}
	}
}

func (m *manifest) toType(t types.Event) reflect.Type {
	var v = reflect.ValueOf(t)
	if reflect.Ptr == v.Kind() || reflect.Interface == v.Kind() {
		v = v.Elem()
	}
	return v.Type()
}
