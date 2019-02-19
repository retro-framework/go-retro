package commands

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gobuffalo/flect"
	"github.com/retro-framework/go-retro/framework/retro"
)

func NewManifest() retro.CommandManifest {
	return &manifest{
		m: make(map[reflect.Type][]retro.Command),
		a: make(map[retro.Command]reflect.Type),
	}
}

var DefaultManifest = NewManifest()

func Register(agg retro.Aggregate, cmd retro.Command) error {
	return DefaultManifest.Register(agg, cmd)
}

func RegisterWithArgs(agg retro.Aggregate, cmd retro.Command, arg retro.CommandArgs) error {
	return DefaultManifest.RegisterWithArgs(agg, cmd, arg)
}

type manifest struct {
	// Agg:[]Command
	m map[reflect.Type][]retro.Command
	// Command:Args
	a map[retro.Command]reflect.Type
}

func (m *manifest) Register(agg retro.Aggregate, cmd retro.Command) error {
	if existingCmds, anyCmds := m.m[m.toType(agg)]; anyCmds {
		for _, existingCmd := range existingCmds {
			if cmd == existingCmd {
				return fmt.Errorf("can't register command %s for aggregate %s, command already registered", reflect.TypeOf(cmd), m.toType(agg))
			}
		}
	}
	m.m[m.toType(agg)] = append(m.m[m.toType(agg)], cmd)
	return nil
}

func (m *manifest) RegisterWithArgs(agg retro.Aggregate, cmd retro.Command, arg interface{}) error {
	if err := m.Register(agg, cmd); err != nil {
		return err
	}
	m.a[cmd] = m.toType(arg)
	return nil
}

func (m *manifest) ArgTypeFor(c retro.Command) (retro.CommandArgs, bool) {
	if at, ok := m.a[c]; ok {
		return reflect.New(at).Elem().Addr().Interface(), true
	}
	return nil, false
}

func (m *manifest) ForAggregate(agg retro.Aggregate) ([]retro.Command, error) {
	for key, cmds := range m.m {
		if m.toTypeString(m.toType(agg)) == m.toTypeString(key) {
			return cmds, nil
		}
	}
	return []retro.Command{}, nil
}

func (m *manifest) toType(t interface{}) reflect.Type {
	var v = reflect.ValueOf(t)
	if reflect.Ptr == v.Kind() || reflect.Interface == v.Kind() {
		v = v.Elem()
	}
	return v.Type()
}

func (m *manifest) toTypeString(t reflect.Type) string {
	return strings.Join([]string{t.PkgPath(), t.Name()}, ".")
}

func (m *manifest) List() map[string][]string {
	var r = make(map[string][]string)
	for k, v := range m.m {
		if _, ok := r[k.String()]; !ok {
			r[k.Name()] = []string{}
		}
		for _, cmd := range v {
			var cmdName string
			if t := reflect.TypeOf(cmd); t.Kind() == reflect.Ptr {
				cmdName = "*" + t.Elem().Name()
			} else {
				cmdName = t.Name()
			}
			r[k.Name()] = append(r[k.Name()], flect.Underscore(cmdName))
		}
	}
	return r
}
