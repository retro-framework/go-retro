package main

//
// import (
// 	"fmt"
// 	"path"
// 	"strings"
//
// 	"github.com/retro-framework/go-retro/aggregates"
// 	"github.com/retro-framework/go-retro/storage"
// 	"github.com/pkg/errors"
// )
//
// type AggregateFactory func() (aggregates.Aggregate, error)
// type CmdFactoryMap map[string]func(aggregates.Aggregate) aggregates.Command
//
// type AggregateConf struct {
// 	factory  AggregateFactory
// 	commands CmdFactoryMap
// }
//
// type Resolver struct {
// 	conf map[string]AggregateConf
// 	// aggregateFactories  map[string]func() (Aggregate, error)
// 	// aggregateCommandMap map[string]CmdFactoryMap
// }
//
// // Register is used to tell the resolver about a pathname to associate with a type.
// // because go doesn't have first class types we need to give it a factory function
// // which can be used to get an instance of a type and then rehydrate it.
// func (r *Resolver) Register(aggPath string, factory func() (aggregates.Aggregate, error), cmds CmdFactoryMap) error {
//
// 	// Check that aggregateFactories is initialized
// 	if r.conf == nil {
// 		r.conf = map[string]AggregateConf{}
// 	}
//
// 	r.conf[aggPath] = AggregateConf{factory, cmds}
//
// 	// Ensure that they haven't passed something like users/bananas/ (is that a problem?)
// 	// if parts := strings.Split(aggPath, "/"); len(parts) > 1 {
// 	// 	return errors.New("can't register aggPath that ")
// 	// }
//
// 	// Ensure that we don't have any dir separators, because we use the paths library that
// 	// expects forward slash delimiters we should go from there.
// 	// > The path package should only be used for paths separated by forward
// 	// > slashes, such as the paths in URLs. This package does not deal with
// 	// > Windows paths with drive letters or backslashes; to manipulate operating
// 	// > system paths, use the path/filepath package.
// 	//
// 	// (is this check redundant if we're checking above for num parts?)
// 	// if ans := strings.ContainsAny(dirname, "/"); ans {
// 	// 	return errors.New("dirname may not include a slash")
// 	// }
//
// 	// Don't allow accidental overwriting of a handler, if this should be a use-case
// 	// one day we should provide a de-/re-register function for that.
// 	// if _, ok := r.aggregateFactories[dirname]; ok {
// 	// 	return errors.New("handler already defined, failing")
// 	// }
//
// 	return nil
//
// }
//
// // canonicalize ensures that a path component is given, by
// func (r *Resolver) canonicalize(cmd CmdDesc) (CmdDesc, error) {
// 	var err error
// 	ns, fn := path.Split(cmd.Name())
// 	if ns == "" || ns == "_" {
// 		return Cmd{path.Join("_", fn), cmd.Args()}, err
// 	}
// 	return Cmd{path.Join(strings.ToLower(ns), fn), cmd.Args()}, nil
// }
//
// func (r *Resolver) Resolve(repo storage.Depot, cmd CmdDesc) (aggregates.TransFn, error) {
//
// 	// Check we have a repository
// 	if repo == nil {
// 		return nil, errors.New("c")
// 	}
//
// 	// Ensure command path has been normalized
// 	cmd, err := r.canonicalize(cmd)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "can't canonicalize CmdDesc")
// 	}
//
// 	// Look up the factory function, we'll need this to get an empty
// 	// instance for the repository to rehydrate for us.
// 	//
// 	// TODO: should we use the type checking here to warn people if
// 	// their aggregate factories don't return pointers?
// 	aggConf, found := r.conf[cmd.Path()]
// 	if !found {
// 		return nil, errors.Errorf("could not find handler at path %s (%s)", cmd.Path(), cmd)
// 	}
// 	aggFactory := aggConf.factory
//
// 	agg, err := aggFactory()
// 	if err != nil {
// 		return nil, errors.Errorf("could not initialize aggregate factory %s (%s)", cmd.Path(), cmd)
// 	}
//
// 	err = repo.Rehydrate(agg, cmd.Path()) // faking rehydration with repo.go:45 mocks
// 	if err != nil {
// 		return nil, errors.Wrap(err, fmt.Sprintf("can't lookup aggregate %s 404", cmd.Path()))
// 	}
//
// 	method := path.Base(cmd.Name())
// 	if cmdFactory, found := aggConf.commands[method]; found {
// 		return cmdFactory(agg).Apply, nil
// 	} else {
// 		return nil, errors.Errorf("couldn't find %s in %v", method, aggConf.commands)
// 	}
//
// }
