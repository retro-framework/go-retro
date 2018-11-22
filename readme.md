
[![Build Status](https://travis-ci.org/retro-framework/go-retro.svg?branch=master)](https://travis-ci.org/retro-framework/go-retro) [![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/retro-framework/go-retro)

# Retro Framework - Go-Retro

    retrospective
    rɛtrə(ʊ)ˈspɛktɪv

    > adjective
    > 1. looking back on or dealing with past events or situations.
    > "our survey was retrospective"
    > synonyms:	backdated, retroactive, ex post facto, backward-looking

## Usage

    $ git clone https://github.com/retro-framework/go-retro.git 
    $ cd go-retro
    $ go test ./framework/...

**NOTE:** Some tests (integration, external server) are guarded by build tags, build with:

    $ go test -tags 'redis integration' ./framework/...
  
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

The doctor executable needs to check for the following:

- Command fn SetState having a non pointer receiver (defer to type system?)
- Aggregates having non pointer receivers (defer to type system?)
- Events having non-exported fields (as we can't inspect on them)
- Unmounted events, aggregates and commands (warning, not an error)
- Session start function that returns no event (static analysis?)

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

## Example Packed Checkpoint

    $ cat test.file
    checkpoint 120affix sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
    session DEADBEEF-SESSIONID

    {"foo":"bar"}

Hex dumping it shows the null byte (00) between the `checkpoint 120` and `affix`:

    $ hexdump -C -n128  test.file
    00000000  63 68 65 63 6b 70 6f 69  6e 74 20 31 32 30 00 61  |checkpoint 120.a|
    00000010  66 66 69 78 20 73 68 61  32 35 36 3a 32 63 66 32  |ffix sha256:2cf2|
    00000020  34 64 62 61 35 66 62 30  61 33 30 65 32 36 65 38  |4dba5fb0a30e26e8|
    00000030  33 62 32 61 63 35 62 39  65 32 39 65 31 62 31 36  |3b2ac5b9e29e1b16|
    00000040  31 65 35 63 31 66 61 37  34 32 35 65 37 33 30 34  |1e5c1fa7425e7304|
    00000050  33 33 36 32 39 33 38 62  39 38 32 34 0a 73 65 73  |3362938b9824.ses|
    00000060  73 69 6f 6e 20 44 45 41  44 42 45 45 46 2d 53 45  |sion DEADBEEF-SE|
    00000070  53 53 49 4f 4e 49 44 0a  0a 7b 22 66 6f 6f 22 3a  |SSIONID..{"foo":|
    00000080

