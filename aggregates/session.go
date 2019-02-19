package aggregates

import (
	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/retro"
)

type Session struct {
	NamedAggregate

	HasIdentity  bool
	IdentityName retro.PartitionName
}

func (agg *Session) ReactTo(aev retro.Event) error {
	switch ev := aev.(type) {
	case *events.StartSession:
	case *events.AssociateIdentity:
		if ev.Identity != nil {
			agg.HasIdentity = true
			agg.IdentityName = ev.Identity.Name()
		}
	default:
		return errors.Errorf("Session aggregate doesn't know what to do with %s", aev)
	}
	return nil
}

func init() {
	Register("session", &Session{})
}
