package aggregates

import (
	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/types"
)

type WidgetsApp struct {
	AllowCreateIdentities                  bool
	AllowCreateAuthorizations              bool
	AllowBindingIdentitiesToAuthorizations bool
}

func (wa *WidgetsApp) ReactTo(ev types.Event) error {
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
