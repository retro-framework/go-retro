package resolver

import (
	"encoding/json"
	"fmt"

	"github.com/retro-framework/go-retro/framework/types"
)

type resolver struct {
	aggm types.AggregateManifest
	cmdm types.CommandManifest
}

func New(aggm types.AggregateManifest, cmdm types.CommandManifest) types.Resolver {
	return &resolver{aggm, cmdm}
}

// Resolve uses the byte slice provided and unmarshals it with JSON
// the provided byte slice must at least name name and path. The JSON
// object may also contain an "args" key which will be used for a second
// phase of unmarshalling after using the path and name to resolve the
// aggregate and command respectively. To construct the command object
// the registered type will be instantiated if the manifest contains a
// type for the args.
func (r *resolver) Resolve(depot types.Depot, b []byte) (types.CommandFunc, error) {

	// We expect a JSON object with at least "name" and "path" this is
	// effectively like "identity/SetProfilePic" if path contains more than
	// one segment we have to do some magic to strip it, at this point
	// however we're not interested in the concrete instance of the aggregate
	// or not, we just need it's type in order to find it's command set and
	// proceed to deconstructing the byte slice into real objects.
	var cmdDesc commandDesc
	err := json.Unmarshal(b, &cmdDesc)
	if err != nil {
		return nil, err // TODO: wrap me
	}

	// Check if the given path corresponds to a known aggregate,
	// if not we might consider falling back to `_` to effectively
	// make the "_" optional upstream.
	agg, err := r.aggm.ForPath(cmdDesc.Path)
	if err != nil {
		return nil, err // TODO: wrap me
	}
	if agg == nil {
		return nil, fmt.Errorf("could not find aggreate")
	}

	// Here we use the given depot to rehydrate the aggregage given
	// a certain path.
	err = depot.Rehydrate(agg, cmdDesc.Path)
	if err != nil {
		return nil, err // TODO: wrap me
	}

	fmt.Printf("Parsed cmdDesc: %#v\n", cmdDesc)
	return nil, nil
}

type commandDesc struct {
	Name string `json:"name"`
	Path string
	Args types.ApplicationCmdArgs
}
