package aggregates

import (
	"fmt"
	"reflect"

	"github.com/retro-framework/go-retro/framework/types"
)

// THIS FILE SHOULD BE AUTO GENERATED FROM FILES IN THE AGGREGATES
// DIRECTORY YOU MAY NOT EDIT IT BY HAND

var DefaultManifest = NewManifest()

func NewManifest() types.AggregateManifest {
	return &manifest{make(map[string]reflect.Type)}
}

// Register takes a Zero Value of an aggreate and stores
// it in the manifest map for later lookup with ForPath
// which will return a copy.
func Register(path string, agg types.Aggregate) error {
	return DefaultManifest.Register(path, agg)
}

type manifest struct {
	m map[string]reflect.Type
}

func (m *manifest) Register(path string, agg types.Aggregate) error {
	if existing, exists := m.m[path]; exists {
		return fmt.Errorf("can't register aggregate %s at %s, path already bound to %s", reflect.TypeOf(agg), path, existing)
	}
	m.m[path] = m.toType(agg)
	return nil
}

// ForPath looks up the reflect.Type registered for the string path given and
// constructs a new zero value of that type using the reflect package.
//
// A pointer to that type is returned cast to the types.Aggregate interface it
// is important that the type given defines its methods with a pointer receiver
// to ensure that types returned by this function correctly implement the
// types.Aggregate interface.
func (m *manifest) ForPath(path string) (types.Aggregate, error) {
	if et, exists := m.m[path]; exists {
		return reflect.New(et).Elem().Addr().Interface().(types.Aggregate), nil
	}
	return nil, nil
}

// toType takes an Aggregate (or pointer to aggregate) and returns it's
// reflect.Type, this type is used to later reconstruct an empty zero valued
// instance of this type when looking up with ForPath
//
// Whilst it is required for correctness that Aggregate defines it's methods
// on it's pointer type, this function accepts both, and unwraps them as
// appropriate.
//
// > Be conservative in what you do, be liberal in what you accept from others
// > – Joe Postel
func (m *manifest) toType(t types.Aggregate) reflect.Type {
	var v = reflect.ValueOf(t)
	if reflect.Ptr == v.Kind() || reflect.Interface == v.Kind() {
		v = v.Elem()
	}
	return v.Type()
}
