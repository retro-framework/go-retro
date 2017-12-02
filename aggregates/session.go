package aggregates

import (
	"github.com/leehambley/ls-cms/framework/types"
	"github.com/pkg/errors"
)

type Session struct {
	IsAnonymous bool
}

func (sesh *Session) ReactTo(ev types.Event) error {
	switch ev {
	default:
		return errors.Errorf("Session aggregate doesn't know what to do with %s", ev)
	}
	return nil
}

func init() {
	Register("session", &Session{true})
}
