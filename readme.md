[![Build Status](https://travis-ci.org/retro-framework/go-retro.svg?branch=master)](https://travis-ci.org/retro-framework/go-retro) [![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/retro-framework/go-retro)

# Retro Framework - Go-Retro

    retrospective
    rɛtrə(ʊ)ˈspɛktɪv

    > adjective
    > 1. looking back on or dealing with past events or situations.
    > "our survey was retrospective"
    > synonyms: backdated, retroactive, ex post facto, backward-looking

Retro is a DDD, CQRS & ES framework. Domain driven design, command query request
segregation and event sourcing.

## Usage

    $ git clone https://github.com/retro-framework/go-retro.git
    $ cd go-retro

Retro is written for Go 1.11+ and does not expect to be loaded in the `GOPATH`.
Trying to use Retro from within the `GOPATH` may have unintended consequences as
`go mod` may not work properly.

## Running Tests

The easy part:

    make test-units
    make test-integration

Testing just a single package:

    make test-units PACKAGE=github.com/retro-framework/go-retro/framework/depot/...

The working directory is mounted at `/go/src/github.com/...` inside the testing
container hence the complete package URL must be used.

To run the tests with a specific set of flags (see `GOTESTFLAGS?=...` in the
`Makefile`) for the defaults. For example, you may wish to run `make GOTESTFLAGS="-v"`
to see verbose output, or specify `-run ...` with a pattern to run only tests for a 
matching a specific pattern.

**NOTE:** Some tests (integration, external server) are guarded by build tags,
build with:

    make test-integration GOTESTTAGS=redis

Extra flags can be provided to `go test` by setting `GOTESTFLAGS` for example:

    make test-units GOTESTFLAGS="-count 1000"

Of course, all combinations are also supported. These are exposed to GNU Make
as environment variables, so they can also be baked into scripts or `export`ed
in your shell.

## Running Development Server

    $ STORAGE_PATH=./depot_common_test519643404 ~/go/bin/gin run demo-server.go

_Note:_ This demo server is relatively incomplete, but it's currently the least
worst way to run Retro.

## Broad Concept:

### The Write Model (**C**QRS)

Retro expects a serialzied command from downstream clients in the form of:

```
{"path":"users/123", "name":"createProfile", "params": {...}}
```

The likely downstream candidate is an HTTP server, so this could easily be
inferred from a POST body such as:

```
POST /users/123/
{"createProfile": { ... }}
```

The serialized form is then used to look up in the "Aggregate" and "Command" manifests
to ensure that an aggregate `User` (mounted as `user(s)`) is registered, and it has a
set of commands that includes `CreateProfile` (`createProfile`).

Assuming that both are found, a zero value user will be created from the prototype in
the manifest and "rehydrated" by replaying events against it.

This is on the "Command" side of the system, so the `User` in this instance is not a
view-object and may discard events that have no business value.

The command object will then be instantiated, and passed the rehydrated `User` and will
have it's `Apply` function called. The `Apply` function may reach out to other models
in the system and will have access to a `Session` (first-class concept) and can make
decisions about business logic as it sees fit.

Assuming succeess the command object should return a `map[PartitionName][]Event`
(`PartitionName` is a type alias for `String`).

If the command returns in this way, a new `Checkpoint` will be added to the storage
containing all of the metadata we had to hand, with the new events all stored neatly
in the storage.

The underlying store is Git-like in that it has a CAS (content addressable storage)
object database which stores the three kinds of objects (checkpoints, events, affixes)
and a ref database which stores pointers to the "newest" on any given "branch".

The storage can be visualized as:

```
                         ▽   - References (refs)
────────◍──────◍──◍──────◍   - Checkpoints
┌─┬─┬─┬─┐┌─┬─┬─┐┌─┐┌─┬─┬─┐   - Affixes
│★│❂│■│✪││★│❂│■││★││★│❂│■│   - Events
└─┴─┴─┴─┘└─┴─┴─┘└─┘└─┴─┴─┘
```

Beginning with the bottom line and the events enclosed in boxes `★,❂,■,✪` represent different
kinds of "Event", these are the historical record of changes made. Examples may include:

- `ProfilePictureChanged`
- `FrienshipRequestAccepted`
- `OAuthTokenGenerated`
- `WelcomeEmailEnqueued`

An `Affix` may contain events for more than one Partition (Aggregate), an
example `CreateUser` command may emit events like this (icons are not part
of the serialization format and are used only for illustrative purposes):

```
    ★ profile/456 SetName          {name: "Paul"}
    ❂ profile/456 SetPicture       {url: "s3://..."}
    ■ profile/456 SetVisibility    {visibility: "public"}
    ✪ users/123   AssociateProfile {profile_urn: "profiles/456"}
```

This asymmetry is intentional and vital to the usability of the framework.
Subsequent `...Profile` Commands may emit the _same_ events, but in shorter
chunks. An `UpdateProfilePicture` Command may emit simply `profile/456 SetPicture {url: "s3://..."}`. The atomicity of events is important to get
right.

The `────────◍` annotation in the visualization illustrates the concept of an `Affix` and
it's relationship to a `Checkpoint`.

A checkpoint serves as a kind of transaction boundary, related events are not live-streamed out
to subscribers, they are "chunked" into affixes and will arrive en-masse on successful command invocations.

A checkpoint has the following structure when unpacked:

```
affix   sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
session DEADBEEF-SESSIONID
date    2019-02-11T14:51:00Z
parent  sha256:5a7d7d9db3bbf8f453a96df99d9cfe1043ef09aa34c44a3171e2ba971387e17f

{"path":"users/123", "name":"createProfile", "params": {...}}
```

The Checkpoint contains the original payload from the downstream client as the "message" (separated from the
headers with two newlines). The "headers" are inspired by Git and HTTP and are simple K/V pairs of metadata.
Some headers such as `date`, `affix` and `session` are mandatory. The headers will usually include
a `parent` key which indicates the previous checkpoint in the history. Checkpoints are in this scheme a singly linked list, or more specifically a Merkle Tree.

In cases (not yet supported) it would be plausible for a checkpoint to have more than one parent to facilitate
merging of parallel histories:

```
                                    ▽
┌─┬─┬─┬─◍─┬───┌─┬─┬─◍──┬──┌─◍─┌─┬─┬─◍
│★│❂│■│✪│ │   │★│❂│■│  │  │★│ │★│❂│■│
└─┴─┴─┴─┘ │   └─┴─┴─┘  │  └─┘ └─┴─┴─┘
          │            │
          │            │
          └──┌─┬─┬─┬─◍─┘
             │★│❂│■│✪│
             └─┴─┴─┴─┘
```

The likely case for this is two concurrent writers conflicting in a way that is safe to temporarily diverge (no
concurrent access to the same shared entities).

The final piece of the puzzle on the "Write" side of the application is the "head pointer" (`▽`). The object DB
contains everything we discussed so far (events, affixes and commands). The refdb contains "pointers" to the newest
checkpoint on a line, or branch which is interesting. By convention Retro keeps the main application state on
`refs/heads/master`.

Upon application of a command all events, affixes and checkpoints are witten to the underlying object database. If
the command was applied _successfully_ however the head pointer for the current branch is moved. This system
has a marginal cost for storing "failed" events (the command payloads are preserved) which are "orphaned" in the
object database, and easily found with some (non-existent) tooling. In the success case, the head pointer move acts
as a transaction boundary guaranteeing that consumers receive batched events which belong together.

Given the following example:

```
                  ▼      ▽
┌─┬─┬─┬─◍┬─┬─┬─◍┬─◍┬─┬─┬─◍
│★│❂│■│✪││★│❂│■││★││★│❂│■│
└─┴─┴─┴─┘└─┴─┴─┘└─┘└─┴─┴─┘
```

Under the hood when the head pointer moves from `▼` to `▽` the consumers will receive three new events without
needing to know about the underlying storage being chunked.

There are a kind of (not yet implemented) head pointer movements which are possible: "non-fast-forward" moves.
These occur when checkpoint now pointed to by the head pointer is not a decendent of the previously indicated checkpoint.

This could happen if the underlying database were modified (which is to be a supported case) and most consumers
would probably benefit from time travelling back to the first event and rebuilding their projection from t0.

```
// start a server:
var e = engine.New(depot, resolveFn, idFn, aggregates.NewManifest(), eventManifest)

// handle a request (get a session, mandatory)
sid, \_ := e.StartSession(req.Context()) // req = http.Request

// try and apply the "dummyCommand" with the paylaod "agg/123"
resStr, err := e.Apply(ctx, sid, []byte(`{"path":"agg/123", "name":"dummyCmd", "params": {...}}`))
```

That `e.Apply` will parse the 3rd arg and go lookup the Command (service object) registered as `dummyCmd`, instantiate
it with a rehydrated `agg/123` (rehydration is the "write" sides version of a projection, events from the depot are
replayed and anything pertinent to the service object can be tracked, incase agg/123 is new, then a Zero Value instance
will be made.)

Assuming successful return from DummyCmd(), the return value should contain a slice of new events.

The engine then takes those events, packs and stores them in the odb, it the makes an affix with the event hash IDs and
the names of the partitions to which they belong, and packs and stores _that_. It finally then makes a checkpoint, the
checkpoint contains all the request metadata; original payload, session ID, date and time, etc, etc and packs and stores
that too.

Assuming all was well until now, the head pointer moves by using the depot API to notify subscribers that hte head moved
fromX toY which causes the new events, and maybe new partition to propagate to the subscribers on the read side.

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

```

```
