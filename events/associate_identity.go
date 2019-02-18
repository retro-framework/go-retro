package events

import "github.com/retro-framework/go-retro/framework/types"

type AssociateIdentity struct {
	Identity types.ExistingAggregate `json:"identity"`
}

func init() {
	Register(&AssociateIdentity{})
}
