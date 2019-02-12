package aggregates

import (
	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/framework/types"
)

type Session struct {
	NamedAggregate
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
	Register("session", &Session{IsAnonymous: true})
}
