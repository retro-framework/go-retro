package events

import "github.com/retro-framework/go-retro/framework/retro"

type AssociateIdentity struct {
	Identity retro.ExistingAggregate `json:"identity"`
}

func init() {
	Register(&AssociateIdentity{})
}
