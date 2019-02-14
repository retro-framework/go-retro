package events

import "github.com/retro-framework/go-retro/framework/types"

type AssociateIdentity struct {
	Identity types.PartitionName `json:"name"`
}

func init() {
	Register(&AssociateIdentity{})
}
