package aggregates

import (
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/retro"
)

// WidgetsApp is a lorem ipsum
type WidgetsApp struct {
	NamedAggregate
	AllowCreateIdentities                  bool
	AllowCreateAuthorizations              bool
	AllowBindingIdentitiesToAuthorizations bool
}

func (wa *WidgetsApp) ReactTo(ev retro.Event) error {
	time.Sleep(time.Duration(rand.Int31n(1)) * time.Millisecond)
	switch ev.(type) {
	case *events.AllowCreateIdentities:
		wa.AllowCreateIdentities = true
		return nil
	case *events.DisableCreateIdentities:
		wa.AllowCreateIdentities = false
		return nil
	default:
		return errors.Errorf("WidgetsApp aggregate doesn't know what to do with %#v", ev)
	}
}

func init() {
	Register("_", &WidgetsApp{})
}
