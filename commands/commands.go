package commands

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/retro-framework/go-retro/framework/types"
)

func NewManifest() types.CommandManifest {
	return &manifest{make(map[string][]types.Command)}
}

var DefaultManifest = NewManifest()

func Register(agg types.Aggregate, cmd types.Command) error {
	return DefaultManifest.Register(agg, cmd)
}

type manifest struct {
	m map[string][]types.Command
}

func typeToKey(t reflect.Type) string {
	// return strings.Join([]string{filepath.Base(t.PkgPath()), t.Name()}, ".")
	return strings.Join([]string{t.Name()}, ".")
}

func (m *manifest) Register(agg types.Aggregate, cmd types.Command) error {
	// https://gist.github.com/hvoecking/10772475#file-translate-go-L191
	// contains nicer reflection code with better explanations
	var aggType = reflect.TypeOf(agg)
	if existingCmds, anyCmds := m.m[typeToKey(aggType)]; anyCmds {
		for _, existingCmd := range existingCmds {
			if cmd == existingCmd {
				return fmt.Errorf("can't register command %s for aggregate %s, command already registered", reflect.TypeOf(cmd), aggType)
			}
		}
	}
	m.m[typeToKey(aggType)] = append(m.m[typeToKey(aggType)], cmd)
	return nil
}

func (m *manifest) ForAggregate(agg types.Aggregate) ([]types.Command, error) {
	// https://gist.github.com/hvoecking/10772475#file-translate-go-L191
	// contains nicer reflection code with better explanations
	var aggType = reflect.TypeOf(agg)
	return m.m[typeToKey(aggType)], nil
}
