package aggregates

import (
	"fmt"
	"reflect"

	"github.com/retro-framework/go-retro/framework/types"
)

// THIS FILE SHOULD BE AUTO GENERATED FROM
// FILES IN THE EVENTS DIRECTORY YOU MAY
// NOT EDIT IT BY HAND

var DefaultManifest = NewManifest()

func NewManifest() types.AggregateManifest {
	return &manifest{make(map[string]reflect.Type)}
}

func Register(path string, agg types.Aggregate) error {
	return DefaultManifest.Register(path, agg)
}

type manifest struct {
	m map[string]reflect.Type
}

func (m *manifest) Register(path string, agg types.Aggregate) error {
	if existing, exists := m.m[path]; exists {
		return fmt.Errorf("Can't register aggregate %s at %s, path already bound to %s", reflect.TypeOf(agg), path, existing)
	}
	m.m[path] = reflect.TypeOf(agg)
	return nil
}

func (m *manifest) ForPath(path string) (types.Aggregate, error) {
	return nil, nil
}
