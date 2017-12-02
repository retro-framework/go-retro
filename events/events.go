package events

import (
	"fmt"
	"reflect"

	"github.com/leehambley/ls-cms/framework/types"
)

var Manifest = map[string]reflect.Type{}

func Register(ev types.Event) error {
	// TODO: does this only accept pointers to evs because of .Elem().Name() ???
	//       see skipped test
	evName := reflect.TypeOf(ev).Elem().Name()
	if _, exists := Manifest[evName]; exists {
		return fmt.Errorf("Can't register event %s, already registered", evName)
	}
	Manifest[evName] = reflect.TypeOf(ev)
	return nil
}
