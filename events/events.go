package events

import (
	"fmt"
	"reflect"

	"github.com/retro-framework/go-retro/framework/types"
)

var Manifest = map[string]reflect.Type{}

func Register(ev types.Event) error {
	// https://gist.github.com/hvoecking/10772475#file-translate-go-L191
	// contains nicer reflection code with better explanations
	//
	// TODO: does this only accept pointers to evs because of .Elem().Name() ???
	//       see skipped test
	evName := reflect.TypeOf(ev).Elem().Name()
	if _, exists := Manifest[evName]; exists {
		return fmt.Errorf("can't register event %s, already registered", evName)
	}
	Manifest[evName] = reflect.TypeOf(ev)
	return nil
}
