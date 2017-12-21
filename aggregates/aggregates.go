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
	return &manifest{make(map[string]interface{})}
}

// Register takes a Zero Value of an aggreate and stores
// it in the manifest map for later lookup with ForPath
// which will return a copy.
func Register(path string, agg types.Aggregate) error {
	return DefaultManifest.Register(path, agg)
}

type manifest struct {
	m map[string]interface{}
}

func (m *manifest) Register(path string, agg types.Aggregate) error {
	if existing, exists := m.m[path]; exists {
		return fmt.Errorf("Can't register aggregate %s at %s, path already bound to %s", reflect.TypeOf(agg), path, existing)
	}
	m.m[path] = agg
	return nil
}

func (m *manifest) ForPath(path string) (types.Aggregate, error) {
	if existing, exists := m.m[path]; exists {
		rv := reflect.TypeOf(existing)
		for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
			rv = rv.Elem()
		}
		ms := reflect.New(rv).Elem().Interface()
		return ms.(types.Aggregate), nil
	}
	return nil, nil
}
