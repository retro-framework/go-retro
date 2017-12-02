package commands

import (
	"fmt"
	"reflect"

	"github.com/leehambley/ls-cms/framework/types"
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
	var aggType = reflect.TypeOf(agg)
	if existingCmds, anyCmds := m.m[aggType]; anyCmds {
		for _, existingCmd := range existingCmds {
			if cmd == existingCmd {
				return fmt.Errorf("Can't register command %s for aggregate %s, command already registered", reflect.TypeOf(cmd), aggType)
			}
		}
	}
	m.m[aggType] = append(m.m[aggType], cmd)
	return nil
}

func (m *manifest) ForAggregate(types.Aggregate) ([]types.Command, error) {
	return nil, nil
}
