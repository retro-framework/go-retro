package aggregates

import (
	"github.com/leehambley/ls-cms/events"
	"github.com/pkg/errors"
)

type Session struct {
	isAnonymous bool
}

func (sesh *Session) ReactTo(ev events.Event) error {
	switch ev {
	default:
		return errors.Errorf("Session aggregate doesn't know what to do with %s", ev)
	}
	return nil
}

func (sesh *Session) IsAnonymous() bool {
	return true
}

func init() {
	Register("session", &Session{})
}
