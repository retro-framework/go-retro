package resolver

import (
	"context"
	"encoding/json"
	"errors"
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
	spnUnmarshal := opentracing.StartSpan("unmarshal json byte steam", opentracing.ChildOf(spnResolve.Context()))
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
		return nil, Error{"parse-agg-path", fmt.Errorf("agg path %q does not include an id", cmdDesc.Path)}
	} else if len(aggType) == 0 && len(aggID) > 0 { // no `/` in path
		// TODO: Check for this case in the test
	}

	// Check if the given path corresponds to a known aggregate,
	// if not we might consider falling back to `_` to effectively
	// make the "_" optional upstream.
	spnAggLookup := opentracing.StartSpan("look up aggregate", opentracing.ChildOf(spnResolve.Context()))
	defer spnAggLookup.Finish()
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
	spnAggCmdLookup := opentracing.StartSpan("look up command for aggregate", opentracing.ChildOf(spnResolve.Context()))
	defer spnAggCmdLookup.Finish()
	cmds, err := r.cmdm.ForAggregate(agg)
	spnAggCmdLookup.SetTag("commands.list", cmds)
	spnAggCmdLookup.SetTag("commands.num", len(cmds))
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
		sp.LogKV(reflect.TypeOf(c).Elem().Name(), cmdDesc.Name)
		if strings.Compare(reflect.TypeOf(c).Elem().Name(), cmdDesc.Name) == 0 {
			sp.LogKV("matched", reflect.TypeOf(c).Elem().Name())
			cmd = c
		} else {
			sp.LogKV("no match", reflect.TypeOf(c).Elem().Name())
		}
	}
	sp.Finish()

	if cmd == nil {
		return nil, Error{"agg-cmd-lookup", fmt.Errorf("no command registered with name %s for aggregate %v", cmdDesc.Name, reflect.TypeOf(agg).Elem().Name())}
	}

	fmt.Println("about to rehydrate")
	err = depot.Rehydrate(ctx, agg, types.PartitionName(cmdDesc.Path))
	if err != nil {
		return nil, Error{"agg-rehydrate", err}
	}
	fmt.Println("…done")

	cmd.SetState(agg)

	if len(cmdDesc.Args) > 0 {
		if cmdWithArgs, ok := cmd.(types.CommandWithArgs); !ok {
			return nil, Error{"cast-cmd-with-args", errors.New("args given, but command does not implement CommandWithArgs")}
		} else {
			if err := cmdWithArgs.SetArgs(cmdDesc.Args); err != nil {
				return nil, Error{"assign-args", err}
			}
		}
	}

	// TODO: Could implement an INFO level warning incase args are absent but
	//       the command actually implements CommandWithArgs (annoying if use-
	//			 case permits optional args?)

	return cmd.Apply, nil
}

type commandDesc struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Args types.CommandArgs
}
