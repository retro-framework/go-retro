package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/retro-framework/go-retro/framework/types"
)

type Error struct {
	Op  string
	Err error
}

func (e Error) Error() string {
	return fmt.Sprintf("resolver: op: %q err: %q", e.Op, e.Err)
}

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
func (r *resolver) Resolve(ctx context.Context, depot types.Depot, b []byte) (types.CommandFunc, error) {

	spnResolve, ctx := opentracing.StartSpanFromContext(ctx, "resolver.Resolve")
	defer spnResolve.Finish()

	// We expect a JSON object with at least "name" and "path" this is
	// effectively like "identity/SetProfilePic" if path contains more than
	// one segment we have to do some magic to strip it, at this point
	// however we're not interested in the concrete instance of the aggregate
	// or not, we just need it's type in order to find it's command set and
	// proceed to deconstructing the byte slice into real objects.
	spnUnmarshal, ctx := opentracing.StartSpanFromContext(ctx, "unmarshal JSON byte stream")
	spnUnmarshal.SetTag("payload", string(b))
	defer spnUnmarshal.Finish()
	var cmdDesc commandDesc
	err := json.Unmarshal(b, &cmdDesc)
	if err != nil {
		err = Error{"json-unmarshal", err}
		spnUnmarshal.LogKV("event", "error", "error.object", err)
		return nil, err
	}
	spnUnmarshal.Finish()

	// Check if the aggregate described by the "path" field of the cmdDesc
	// includes "basename"
	cmdDescParts := strings.Split(strings.TrimSpace(cmdDesc.Path), "/")
	if len(cmdDescParts) > 2 {
		return nil, Error{"parse-agg-path", fmt.Errorf("agg path %q contains too many slashes (may not nest)", cmdDesc.Path)}
	}

	if len(cmdDescParts) < 2 {
		return nil, Error{"parse-agg-path", fmt.Errorf("agg path %q does not split into exactly two parts", cmdDesc.Path)}
	}

	aggType, aggID := cmdDescParts[0], cmdDescParts[1]
	spnUnmarshal.SetTag("agg.type", aggType)
	spnUnmarshal.SetTag("agg.id", aggID)

	if len(aggType) == 0 && len(aggID) == 0 { // neither aggName or ID given … maybe we route to `_` if defined?
		// TODO: Check if there's a "_" aggregate defined (may likely not be the case in many tests)
		return nil, Error{"parse-agg-path", fmt.Errorf("can't split %q into name and id, both parts empty (empty string?)", cmdDesc.Path)}

	} else if len(aggType) > 0 && len(aggID) == 0 { // path given, no ID

	} else if len(aggType) == 0 && len(aggID) > 0 { // no `/` in path
		return nil, Error{"parse-agg-path", fmt.Errorf("agg path %q does not include an id", cmdDesc.Path)}
	}

	// Check if the given path corresponds to a known aggregate,
	// if not we might consider falling back to `_` to effectively
	// make the "_" optional upstream.
	spnAggLookup, ctx := opentracing.StartSpanFromContext(ctx, "lookup aggregate")
	defer spnAggLookup.Finish()
	fmt.Println("aggType", aggType)
	agg, err := r.aggm.ForPath(aggType)
	if err != nil {
		err = Error{"agg-lookup", err}
		spnAggLookup.LogKV("event", "error", "error.object", err)
		return nil, err
	}
	if agg == nil {
		err := Error{"agg-lookup", fmt.Errorf("could not find aggreate")}
		spnAggLookup.LogKV("event", "error", "error.object", err)
		return nil, err
	}
	spnAggLookup.Finish()

	// Look up the command before we invest effort to rehydrate something
	// we might not be able to use
	spnAggCmdLookup, ctx := opentracing.StartSpanFromContext(ctx, "lookup aggregate command")
	defer spnAggCmdLookup.Finish()
	cmds, err := r.cmdm.ForAggregate(agg)
	if err != nil {
		err = Error{"agg-cmd-lookup", err}
		spnAggCmdLookup.LogKV("event", "error", "error.object", err)
		return nil, err
	}

	if len(cmds) == 0 {
		return nil, Error{"agg-cmd-lookup", fmt.Errorf("aggregate has no registered commands")}
	}

	sp := opentracing.StartSpan("iterating over aggregate commands", opentracing.ChildOf(spnAggCmdLookup.Context()))
	var cmd types.Command
	for _, c := range cmds {
		if strings.Compare(reflect.TypeOf(c).Elem().Name(), cmdDesc.Name) == 0 {
			cmd = c
		} else {
			sp.LogKV("no match", reflect.TypeOf(c).Elem().Name())
		}
	}
	sp.Finish()

	if cmd == nil {
		return nil, Error{"agg-cmd-lookup", fmt.Errorf("no command registered with name %s for aggregate %v", cmdDesc.Name, reflect.TypeOf(agg).Elem().Name())}
	}

	// TODO: ~~instrument this and~~ make sure it works in general!
	err = depot.Rehydrate(ctx, agg, cmdDesc.Path)
	if err != nil {
		return nil, Error{"agg-rehydrate", err}
	}

	return cmd.Apply, nil
}

type commandDesc struct {
	Name string `json:"name"`
	Path string
	Args types.ApplicationCmdArgs
}
