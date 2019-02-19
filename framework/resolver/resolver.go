package resolver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gobuffalo/flect"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/retro-framework/go-retro/framework/retro"
)

type Error struct {
	Op  string
	Err error
}

func (e Error) Error() string {
	return fmt.Sprintf("resolver: op: %q err: %q", e.Op, e.Err)
}

type resolver struct {
	aggm retro.AggregateManifest
	cmdm retro.CommandManifest
}

func New(aggm retro.AggregateManifest, cmdm retro.CommandManifest) retro.Resolver {
	return &resolver{aggm, cmdm}
}

// Resolve uses the byte slice provided and unmarshals it with JSON
// the provided byte slice must at least have the key "name".
//
// Optional fields "path" and "args" can be set on the []byte to specify
// that the command targets a "real" aggregate and not the root aggregate
// the args will be conditionally parsed if the command implements retro.CommentWithArgs
//
// Valid values for []byte here include:
//
// {"name":"any_command"} - this will be routed to the Aggregate mounted at "_". Such
// commands may be used to create "root" level objects, in so far as a hierarchy has meaning
// in this part of the model, or to toggle global settings early in the application
// lifecycle.
//
// {"name":"some_command", "path":"user/123"} - this will be routed to Aggregate "user" and
// the repository will be searched for events pertaining to this aggregate name. If the aggregate
// does not exist or has no history to date an error is returned. Assuming that a matching
// aggregate can be rehydrated from the given partitionName then it will be set into the aggregate
// via the command's "SetState" method.
//
// {"name":"some_command", "path":"user/123", "args":{"foo":"bar"}} - as above with args. For
// this to work the command registered under that name must implement retro.CommandWithArgs. The
// arguments will be parsed into a copy of the registered arg type for this command and passed to
// the command the command's "SetArgs" method.
func (r *resolver) Resolve(ctx context.Context, repository retro.Repo, b []byte) (retro.Command, error) {

	spnResolve, ctx := opentracing.StartSpanFromContext(ctx, "resolver.Resolve")
	defer spnResolve.Finish()

	// We expect a JSON object with at least "name" and "path" this is
	// effectively like "identity/SetProfilePic" if path contains more than
	// one segment we have to do some magic to strip it, at this point
	// however we're not interested in the concrete instance of the aggregate
	// or not, we just need it's type in order to find it's command set and
	// proceed to deconstructing the byte slice into real objects.
	spnUnmarshal := opentracing.StartSpan("unmarshal command description", opentracing.ChildOf(spnResolve.Context()))
	spnUnmarshal.SetTag("payload", string(b))
	var cmdDesc commandDesc
	if err := json.Unmarshal(b, &cmdDesc); err != nil {
		err = Error{"json-unmarshal", err}
		spnUnmarshal.LogKV("event", "error", "error.object", err)
		spnUnmarshal.Finish()
		return nil, err
	}
	spnUnmarshal.Finish()

	// Validate the command description, see the implementation for
	// details
	spnValidateCmdDesc := opentracing.StartSpan("validate command description", opentracing.ChildOf(spnResolve.Context()))
	if errs, ok := cmdDesc.HasErrors(); !ok {
		spnValidateCmdDesc.Finish()
		return nil, Error{"validate-cmd-desc", errs[0]}
	}
	spnValidateCmdDesc.Finish()

	return r.resolve(ctx, spnResolve, repository, cmdDesc)
}

// Resolve does the heavy lifting of finding out what commands and aggregates
// are targeted here.
//
// It may make sense to expose this as a public API to avoid the serialization
// overhead of JSON someday.
func (r *resolver) resolve(ctx context.Context, spnResolve opentracing.Span, repository retro.Repo, cmdDesc commandDesc) (retro.Command, error) {

	// cmdDesc handles the details for us, we may fall-back
	// to looking up "_" if no path was specified.
	spnAggLookup := opentracing.StartSpan("look up aggregate", opentracing.ChildOf(spnResolve.Context()))
	var agg, err = r.aggm.ForPath(cmdDesc.AggregateType())
	if err != nil {
		err = Error{"agg-lookup", err}
		spnAggLookup.SetTag("error", err)
		spnAggLookup.Finish()
		return nil, err
	}
	if agg == nil {
		err := Error{"agg-lookup", fmt.Errorf("could not find aggreate")}
		spnAggLookup.SetTag("error", err)
		spnAggLookup.Finish()
		return nil, err
	}
	spnAggLookup.Finish()

	var setAggregateName = func(a retro.Aggregate, pn retro.PartitionName) error {
		agg.SetName(retro.PartitionName(cmdDesc.Path))
		if err != nil {
			return Error{"agg-assign-name", fmt.Errorf("could not set name on Aggregate: %s", err)}
		}
		if agg.Name() != retro.PartitionName(cmdDesc.Path) {
			return Error{"agg-read-back-name", fmt.Errorf("name change on Aggregate didn't take (check for pointer receivers?)")}
		}
		return nil
	}

	// fmt.Println("checking for existence of cmdDesc.Path", cmdDesc.Path)

	// If the aggregate we're dealing with actually exists then we need to make a few
	// more quick steps... set it's name, and then actually rehydrate it.
	if repository.Exists(ctx, retro.PartitionName(cmdDesc.Path)) {
		if err := setAggregateName(agg, retro.PartitionName(cmdDesc.Path)); err != nil {
			return nil, err
		}
	}

	// return nil, Error{"find-existing-aggregate", fmt.Errorf("no existing aggregate with name: %s", cmdDesc.Path)}
	// Set the aggregate name (useful to ensure that things survive a roundtrip to Commands
	// and back into the Engine)

	// Look up the command before we invest effort to rehydrate something
	// we might not be able to use
	spnAggCmdLookup := opentracing.StartSpan("look up command for aggregate", opentracing.ChildOf(spnResolve.Context()))
	defer spnAggCmdLookup.Finish()
	cmds, err := r.cmdm.ForAggregate(agg)
	spnAggCmdLookup.SetTag("commands.list", cmds)
	spnAggCmdLookup.SetTag("commands.num", len(cmds))
	if err != nil {
		err = Error{"agg-cmd-lookup", err}
		spnAggCmdLookup.SetTag("error", err)
		return nil, err
	}

	if len(cmds) == 0 {
		return nil, Error{"agg-cmd-lookup", fmt.Errorf("aggregate has no registered commands")}
	}

	sp := opentracing.StartSpan("iterating over aggregate commands", opentracing.ChildOf(spnAggCmdLookup.Context()))
	var cmd retro.Command
	for _, c := range cmds {
		var cmdDescName = flect.Pascalize(cmdDesc.Name)
		sp.LogKV(reflect.TypeOf(c).Elem().Name(), cmdDescName)
		// Accept SomeCommand or some_command
		if strings.Compare(reflect.TypeOf(c).Elem().Name(), cmdDescName) == 0 ||
			strings.Compare(reflect.TypeOf(c).Elem().Name(), cmdDesc.Name) == 0 {
			sp.LogKV("matched", reflect.TypeOf(c).Elem().Name())
			cmd = c
		} else {
			sp.LogKV("no match", reflect.TypeOf(c).Elem().Name())
		}
	}
	sp.Finish()
	spnAggCmdLookup.Finish()

	if cmd == nil {
		return nil, Error{"agg-cmd-lookup", fmt.Errorf("no command registered with name %s for aggregate %v", cmdDesc.Name, reflect.TypeOf(agg).Elem().Name())}
	}

	if len(cmdDesc.Args) > 0 {
		var cmdWithArgs, ok = cmd.(retro.CommandWithArgs)
		if !ok {
			return nil, Error{"cast-cmd-with-args", errors.New("args given, but command does not implement CommandWithArgs")}
		}

		var typedArgs, found = r.cmdm.ArgTypeFor(cmd)
		if !found {
			return nil, Error{"assign-args", fmt.Errorf("no arg type registered for cmd, was registered with Register not RegisterWithArgs?")}
		}

		err := json.Unmarshal(cmdDesc.Args, typedArgs)
		if err != nil {
			return nil, Error{"assign-args", err}
		}

		if err := cmdWithArgs.SetArgs(typedArgs); err != nil {
			return nil, Error{"assign-args", err}
		}
	}

	if repository.Exists(ctx, retro.PartitionName(cmdDesc.Path)) {
		spnRehydrate := opentracing.StartSpan("rehydrate target aggregate", opentracing.ChildOf(spnResolve.Context()))
		defer spnRehydrate.Finish()
		err = repository.Rehydrate(ctx, agg, retro.PartitionName(cmdDesc.Path))
		if err != nil {
			// TODO: This exit condition is a nasty "magic string"
			// artefact. It is designed ot match against a string
			// in the "simple-aggregate-rehydrater.go" file where
			// a "not found" error manifests as unknown ref. We don't
			// necessarily expect to find something to rehydrate,
			// this may be a SessionStart event, so we're happy to
			// swallow an error about a non-exixtent partition and
			// failure to rehydrate something we're in the process
			// of creating. These errors need to be better typed.
			if !strings.Contains(err.Error(), "unknown ref") {
				return nil, Error{"agg-rehydrate", err}
			}
		}
	}

	cmd.SetState(agg)

	// TODO: Could implement an INFO level warning incase args are absent but
	//       the command actually implements CommandWithArgs (annoying if use-
	//			 case permits optional args?)

	return cmd, nil
}

type commandDesc struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Args json.RawMessage
}

func (cD commandDesc) HasErrors() ([]error, bool) {
	var errors []error
	if cD.Name == "" {
		errors = append(errors, fmt.Errorf("command name may not be empty"))
	}
	if len(cD.pathParts()) > 2 {
		errors = append(errors, fmt.Errorf("aggregate paths may not contain more than one forwardslash"))
	}
	if len(cD.pathParts()) == 1 && !cD.DoesTargetRootAggregate() {
		errors = append(errors, fmt.Errorf("aggregate path must include an aggregate id after the first forwardslash"))
	}
	return errors, len(errors) == 0
}

func (cD commandDesc) DoesTargetRootAggregate() bool {
	if cD.Path == "_" || cD.Path == "" {
		return true
	}
	return false
}

func (cD commandDesc) AggregateType() string {
	if cD.DoesTargetRootAggregate() {
		return "_"
	}
	return cD.pathParts()[0]
}

func (cD commandDesc) AggregateID() string {
	if cD.DoesTargetRootAggregate() {
		return ""
	}
	return cD.pathParts()[1]
}

func (cD commandDesc) pathParts() []string {
	var r []string
	for _, str := range strings.Split(strings.TrimSpace(cD.Path), "/") {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
