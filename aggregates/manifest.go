package aggregates

import (
	"reflect"
)

// THIS FILE SHOULD BE AUTO GENERATED FROM
// FILES IN THE EVENTS DIRECTORY YOU MAY
// NOT EDIT IT BY HAND

type AggManifest map[string]reflect.Type

var Manifest = AggManifest{}

func Register(path string, agg Aggregate) {
	Manifest[path] = reflect.TypeOf(agg)
}
