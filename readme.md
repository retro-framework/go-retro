
[![Build Status](https://travis-ci.org/retro-framework/go-retro.svg?branch=master)](https://travis-ci.org/retro-framework/go-retro) [![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/retro-framework/go-retro)

# Retro Framework - Go-Retro

    retrospective
    rɛtrə(ʊ)ˈspɛktɪv

    > adjective
    > 1. looking back on or dealing with past events or situations.
    > "our survey was retrospective"
    > synonyms:	backdated, retroactive, ex post facto, backward-looking

A log structured CMS API project from which I hope to extract the log
structured application one day soon.

## Usage

    $ git clone https://github.com/retro-framework/go-retro.git some/path/src/github.com/retro-framework/go-retro
    $ export GOPATH=some/path
    $ brew install dep
    $ (cd some/path/src/github.com/retro-framework/go-retro && dep ensure)
    $ go test github.com/retro-framework/go-retro

**NOTE:** Some tests (integration, external server) are guarded by build tags, build with:

    $ go test -tags 'redis integration' github.com/retro-framework/go-retro/framework/...
  
## Generator

The ls-cms project includes a generator to assist with the creation of the
nonsense boilerplate which is unavoidable.

Usage:

    $ go build .
    $ ls-cms gen aggregate "I'm the name for an aggregate"

## Tests

Testing is grouped into a few areas, aggregates, and whole application stack.

    $ go test .

## Notes:

    type CmdArgs struct {
      // - Apply:
      //     meth: CreateIdentity
      //     args:
      //       name: admin
      //       authorization:
      //         type: EmailAddressWithPassword
      //         args:
      //           username: admin
      //           password: supersecret
      //       role: Unrestricted
    }

## Embedding OpenTracing (Zipkin)

    // import zipkin "github.com/openzipkin/zipkin-go-opentracing"
    //
    // collector, err := zipkin.NewHTTPCollector("http://localhost:9411/api/v1/spans")
	// if err != nil {
	// 	log.Fatal(err)
	// 	return
	// }
	// defer collector.Close()

	// tracer, err := zipkin.NewTracer(
	// 	zipkin.NewRecorder(collector, true, "0.0.0.0:0", "example"),
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// opentracing.SetGlobalTracer(tracer)

	// span, ctx := opentracing.StartSpanFromContext(context.Background(), "Test_Resolver_ResolveExistingCmdSuccessfully")
	// defer span.Finish()

## Doctor

The doctor sub programme needs to check for the following:

- Command fn SetState having a non pointer receiver
- Aggregates having non pointer receivers
- Events having non-exported fields
- Unmounted events, aggregates and commands
- Session start function that returns no event

## Depot

The "depot" comprises a refdb and an odb, these two things are used together to
collectievly store hashed objects and references that point at them.

When a depot is challenged to persist a new command result it should write the
resulting objects to the objdb and update the refdb to point at the new head.

    example
    r := Repository("/tmp")
    ev1 := r.StoreEvent(`{"type":"user:set_name","payload":"Lee"}`)           sha1:4608ae3bceb02c8705734ec5c8a3816efdd489c6
    ev2 := r.StoreEvent(`{"type":"user:set_password","payload":"secret"}`)    sha1:21a4ae47f90ed02b4e310045c05eda7b3e86a050
    ev3 := r.StoreEvent(`{"type":"user:set_email","payload":"lee@lee.com"}`)  sha1:756b61a9d848096d995efea0853ede815e053631
    af1 := r.StoreAffix(Affix{"users/123": [ev1, ev2, ev3]})                  sha1:9d503e2ff8fd91216a8cfb021d1cdbd225b9dfe1
    cp1 := Checkpoint{
             "9d503e2ff8fd91216a8cfb021d1cdbd225b9dfe1",
             `{"path": "users/123", action: "create", "args": {"name":"Lee", .... }}`,
             nil,
             []Checkpoint{},
           }
    cp1h := r.StoreCheckpoint()
    head := r.UpdateRef("heads/master", cp1h)
    tag1 := r.UpdateRef("tags/releaseDate", "message", u)

Heavily inspired by Git.... this "write side" contains all the checkpointing
metadata, etc required to know how a doain ev partition came into being

It also has the nice side-effect that ev objects can be reused to save
space (e.g lots of "accept_tos=1") events can be normalized

I think that there should be a ref object that allows us to have branches
and heads (since commits have parents it would allow us to checkpoint
arbitrary things, and refer to "mainline" parent checkpoints, ala Git
branching models)

Because commits are "branchable" and affixes are bound to partitions
there may be a race condition where you have a stale parent checkpoint
and then fail to be able to commit, but Git solves this by implicitly
making a branch, or having commits with two parents
