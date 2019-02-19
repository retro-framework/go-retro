package aggregates

import (
	"fmt"
	"reflect"

	"github.com/retro-framework/go-retro/framework/retro"
)

// THIS FILE SHOULD BE AUTO GENERATED FROM FILES IN THE AGGREGATES
// DIRECTORY YOU MAY NOT EDIT IT BY HAND

var ErrAggregateNotAnonymous = fmt.Errorf("aggregate: already has a name (is not anonymous)")

type NamedAggregate struct {
	PN retro.PartitionName `json:"name"`
}

func (na *NamedAggregate) Name() retro.PartitionName {
	return na.PN
}

func (na *NamedAggregate) SetName(pn retro.PartitionName) error {
	if len(na.PN) > 0 {
		return ErrAggregateNotAnonymous
	}
	na.PN = pn
	return nil
}

var DefaultManifest = NewManifest()

func NewManifest() retro.AggregateManifest {
	return &manifest{make(map[string]reflect.Type)}
}

// Register takes a Zero Value of an aggreate and stores
// it in the manifest map for later lookup with ForPath
// which will return a copy.
func Register(path string, agg retro.Aggregate) error {
	return DefaultManifest.Register(path, agg)
}

type manifest struct {
	m map[string]reflect.Type
}

func (m *manifest) Register(path string, agg retro.Aggregate) error {
	if existing, exists := m.m[path]; exists {
		return fmt.Errorf("can't register aggregate %s at %s, path already bound to %s", reflect.TypeOf(agg), path, existing)
	}
	m.m[path] = m.toType(agg)
	return nil
}

// ForPath looks up the reflect.Type registered for the string path given and
// constructs a new zero value of that type using the reflect package.
func (m *manifest) ForPath(path string) (retro.Aggregate, error) {
	if et, exists := m.m[path]; exists {
		return reflect.New(et).Elem().Addr().Interface().(retro.Aggregate), nil
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
func (m *manifest) toType(t retro.Aggregate) reflect.Type {
	var v = reflect.ValueOf(t)
	if reflect.Ptr == v.Kind() || reflect.Interface == v.Kind() {
		v = v.Elem()
	}
	return v.Type()
}

func (m *manifest) List() []string {
	var r []string
	for k := range m.m {
		r = append(r, k)
	}
	return r
}
