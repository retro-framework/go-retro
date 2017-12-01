package aggregates

import (
	"fmt"

	"github.com/leehambley/ls-cms/events"
	"github.com/pkg/errors"
)

type WidgetsApp struct {
	allowCreateIdentities                  bool
	allowCreateAuthorizations              bool
	allowBindingIdentitiesToAuthorizations bool
}

func (wa *WidgetsApp) String() string {
	return fmt.Sprintf("App State: wa.allowCreateIdentities: %t", wa.allowCreateIdentities)
}

func (wa *WidgetsApp) ReactTo(ev events.Event) error {
	switch ev.(type) {
	case *events.AllowCreateIdentities:
		wa.allowCreateIdentities = true
		return nil
	case *events.DisableCreateIdentities:
		wa.allowCreateIdentities = false
		return nil
	default:
		return errors.Errorf("WidgetsApp aggregate doesn't know what to do with %#v", ev)
	}
}

func init() {
	Register("_", &WidgetsApp{})
}
