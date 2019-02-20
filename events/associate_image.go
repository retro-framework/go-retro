package events

import (
	"github.com/retro-framework/go-retro/framework/retro"
)

type AssociateImage struct {
	Image retro.ExistingAggregate `json:"image"`
}

func init() {
	Register(&AssociateImage{})
}
