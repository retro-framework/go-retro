package events

import (
	stdlibErrors "errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gobuffalo/flect"
	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/framework/retro"
)

var (
	DefaultManifest = NewManifest()
	ErrNotKnown     = stdlibErrors.New("event not known")
)

type manifest struct {
	m map[string]reflect.Type
}

func NewManifest() retro.EventManifest {
	return &manifest{make(map[string]reflect.Type)}
}

func Register(ev retro.Event) error {
	return DefaultManifest.Register(ev)
}

func (m *manifest) RegisterAs(evName string, ev retro.Event) error {
	var v = m.toType(ev)
	if _, exists := m.m[evName]; exists {
		return fmt.Errorf("can't register event %s, already registered (event names are not namespaced)", evName)
	}
	m.m[evName] = v
	return nil
}

func (m *manifest) Register(ev retro.Event) error {
	var (
		v      = m.toType(ev)
		evName = v.Name()
	)
	return m.RegisterAs(flect.Underscore(evName), ev)
}

func (m *manifest) KeyFor(ev retro.Event) (string, error) {
	for name, tepy := range m.m {
		if tepy == m.toType(ev) {
			return name, nil
		}
	}

	var registeredEvNames []string
	for name := range m.m {
		registeredEvNames = append(registeredEvNames, name)
	}

	// TODO: print a warning here (or include more info in the ErrNotKnown?)
	return "", errors.Wrap(ErrNotKnown, fmt.Sprintf("looking for %q in %q", m.toType(ev), strings.Join(registeredEvNames, ",")))
}

func (m *manifest) ForName(name string) (retro.Event, error) {
	if et, exists := m.m[name]; exists {
		return reflect.New(et).Elem().Addr().Interface().(retro.Event), nil
	}
	return nil, nil
}

func (m *manifest) toType(t retro.Event) reflect.Type {
	var v = reflect.ValueOf(t)
	if reflect.Ptr == v.Kind() || reflect.Interface == v.Kind() {
		v = v.Elem()
	}
	return v.Type()
}

// TODO: This is not tested
func (m *manifest) List() map[string]interface{} {
	var r = make(map[string]interface{})
	for k := range m.m {
		r[k], _ = m.ForName(k)
	}
	return r
}

type WeakURNReference struct {
	U retro.URN
}

func (wur WeakURNReference) URN() retro.URN {
	return wur.U
}
