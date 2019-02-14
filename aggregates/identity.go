package aggregates

import (
	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/types"
)

type Identity struct {
	NamedAggregate

	IsPublic  bool `json:"isPublic"`
	HasAvatar bool `json:"hasAvatar"`
}

func (agg *Identity) ReactTo(aev types.Event) error {
	switch ev := aev.(type) {
	case *events.SetVisibility:
		switch ev.Radius {
		case "public":
			agg.IsPublic = true
		default:
			agg.IsPublic = false
		}
	case *events.SetDisplayName:
	case *events.SetAvatar:
		if len(ev.ImgData) > 0 {
			agg.HasAvatar = true
		}
	default:
		return errors.Errorf("Session aggregate doesn't know what to do with %s", ev)
	}
	return nil
}

func init() {
	Register("identity", &Identity{})
}
