package commands

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/retro-framework/go-retro/framework/types"
)

func NewManifest() types.CommandManifest {
	return &manifest{make(map[reflect.Type][]types.Command)}
}

var DefaultManifest = NewManifest()

func Register(agg types.Aggregate, cmd types.Command) error {
	return DefaultManifest.Register(agg, cmd)
}

type manifest struct {
	m map[reflect.Type][]types.Command
}

func (m *manifest) Register(agg types.Aggregate, cmd types.Command) error {
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

func (m *manifest) ForAggregate(agg types.Aggregate) ([]types.Command, error) {
	for key, cmds := range m.m {
		if m.toTypeString(m.toType(agg)) == m.toTypeString(key) {
			return cmds, nil
		}
	}
	return []types.Command{}, nil
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

// TODO: This is not tested and printing the * for pointers
// without printing the package name is not super helpful
func (m *manifest) List() map[string][]string {
	var r = make(map[string][]string)
	for k, v := range m.m {
		if _, ok := r[k.String()]; !ok {
			r[k.String()] = []string{}
		}
		for _, cmd := range v {
			var cmdName string
			if t := reflect.TypeOf(cmd); t.Kind() == reflect.Ptr {
				cmdName = "*" + t.Elem().Name()
			} else {
				cmdName = t.Name()
			}
			r[k.String()] = append(r[k.String()], cmdName)
		}
	}
	return r
}
