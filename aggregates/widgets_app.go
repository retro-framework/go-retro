package aggregates

import (
	"github.com/leehambley/ls-cms/events"
	"github.com/leehambley/ls-cms/framework/types"
	"github.com/pkg/errors"
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
