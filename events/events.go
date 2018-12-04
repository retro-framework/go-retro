package events

import (
	stdlibErrors "errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/framework/types"
)

var (
	DefaultManifest = NewManifest()
	ErrNotKnown     = stdlibErrors.New("event not known")
)

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

func (m *manifest) KeyFor(ev types.Event) (string, error) {
	for name, tepy := range m.m {
		if tepy == m.toType(ev) {
			return name, nil
		}
	}

	var registeredEvNames []string
	for name, _ := range m.m {
		registeredEvNames = append(registeredEvNames, name)
	}

	// TODO: print a warning here (or include more info in the ErrNotKnown?)
	return "", errors.Wrap(ErrNotKnown, fmt.Sprintf("looking for %q in %q", m.toType(ev), strings.Join(registeredEvNames, ",")))
}

func (m *manifest) ForName(name string) (types.Event, error) {
	if et, exists := m.m[name]; exists {
		return reflect.New(et).Elem().Addr().Interface().(types.Event), nil
	}
	return nil, nil
}

func (m *manifest) toType(t types.Event) reflect.Type {
	var v = reflect.ValueOf(t)
	if reflect.Ptr == v.Kind() || reflect.Interface == v.Kind() {
		v = v.Elem()
	}
	return v.Type()
}
