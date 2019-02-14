package aggregates

import (
	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/types"
)

type Session struct {
	NamedAggregate

	HasIdentity  bool
	IdentityName types.PartitionName
}

func (agg *Session) ReactTo(aev types.Event) error {
	switch ev := aev.(type) {
	case *events.StartSession:
	case *events.AssociateIdentity:
		agg.HasIdentity = true
		agg.IdentityName = ev.Identity
	default:
		return errors.Errorf("Session aggregate doesn't know what to do with %s", aev)
	}
	return nil
}

func init() {
	Register("session", &Session{})
}
