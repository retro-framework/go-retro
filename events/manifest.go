package events

import "reflect"

// THIS FILE SHOULD BE AUTO GENERATED FROM
// FILES IN THE EVENTS DIRECTORY YOU MAY
// NOT EDIT IT BY HAND

type EvManifest map[string]reflect.Type

var Manifest = EvManifest{}

func Register(ev Event) {
	evName := reflect.TypeOf(ev).Elem().Name()
	Manifest[evName] = reflect.TypeOf(ev)
}
