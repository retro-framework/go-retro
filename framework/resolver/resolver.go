package resolver

import (
	"github.com/leehambley/ls-cms/framework/types"
)

type resolver struct {
	aggm types.AggregateManifest
	cmdm types.CommandManifest
}

func New(aggm types.AggregateManifest, cmdm types.CommandManifest) types.Resolver {
	return &resolver{aggm, cmdm}
}

func (r *resolver) Resolve(repo types.Depot, cmd types.CommandDesc) (types.CommandFunc, error) {

	return nil, nil
}
