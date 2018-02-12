package pack

// Used to separate header from payload as a splitter
// character.
//
// BOX DRAWINGS HEAVY QUADRUPLE DASH VERTICAL
// const HeaderContentSepRune = "\u250B"
const HeaderContentSepRune = "\u0000"

// Object is a storable object, they are
// created by
type Object interface {
	TypeName() ObjectTypeName
	Contents() []byte
}

// HashedObject is an object from the hash.
type HashedObject interface {
	Object
	Hash() Hash
}

// Event ns anything serializable for the future
type Event interface{}

// ObjectStore is a generic store for objects. because
// we pack all events, affixes and checkpoints into objects
// which can be serailized we can use an object store to
// store anything. Refs are stored separately.
type ObjectStore interface {
	WriteObject(Object) error
}

// PartitionName as alias for string to make
// the documentation and code examples more
// speaking. Detail should not leak beyond the
// plumbing package.
type PartitionName string

// Affix is a map of partition names to slices of events.
// Affixes are closely related to checkpoints. If a command
// emits a bunch of related events they will be packed into
// a single affix and it will be clear that they were emitted
// at the same time.
//
// An affix *may* be completely empty, or a partition's event
// list may be empty. A failed command execution may yield
// some events, but also an error in which case we would get
// a partial affix, but checkpoint it with an error. The reader
// may prefer to ignore these events, but they do form par
// of our conceptual model.
type Affix map[PartitionName][]Hash

// A checkpoint represents a DDD command object execution
// and persistence of the resulting events. It stores
// an error incase the command failed.
type Checkpoint struct {
	Affix       Hash
	CommandDesc []byte
	Error       error
	Parents     []Checkpoint
}

// Depot stores events, commands, etc. It is heavily inspired
// by Git's model of generic object and ref stores linked with
// pointers. It's aim is to be correct, not fast. To be verifiable,
// and duplicable.
type Depot interface {

	// StoreEvent takes a domain event and returns a Hash
	// the must be deterministic and not affected by PRNG
	// or types, just the serialization format (repository
	// is a storage concern)
	StoreEvent(Event) (Hash, error)

	// StoreAffix stores an affix. An affix may contain
	// a new set of events for one or more partitions, given
	// that we know the name of the aggregate being changed
	// most affixes will contain one partition name and one
	// or more event hashes.
	StoreAffix(Affix) (Hash, error)

	// StoreCheckpoint stores a checkpoint. A checkpoint
	// is approximately equivilant to a Git commit. An
	// object under heavy writes however may "auto branch"
	// checkpoints (multiple checkpoints with a common parent)
	// which we will have to resolve deterministically later
	StoreCheckpoint(Checkpoint) (Hash, error)
}

// example
// r := Repository("/tmp")
// ev1 := r.StoreEvent(`{"type":"user:set_name","payload":"Lee"}`)           sha1:4608ae3bceb02c8705734ec5c8a3816efdd489c6
// ev2 := r.StoreEvent(`{"type":"user:set_password","payload":"secret"}`)    sha1:21a4ae47f90ed02b4e310045c05eda7b3e86a050
// ev3 := r.StoreEvent(`{"type":"user:set_email","payload":"lee@lee.com"}`)  sha1:756b61a9d848096d995efea0853ede815e053631
// af1 := r.StoreAffix(Affix{"users/123": [ev1, ev2, ev3]})                  sha1:9d503e2ff8fd91216a8cfb021d1cdbd225b9dfe1
// cp1 := Checkpoint{
//          "9d503e2ff8fd91216a8cfb021d1cdbd225b9dfe1",
//          `{"path": "users/123", action: "create", "args": {"name":"Lee", .... }}`,
//          nil,
//          []Checkpoint{},
//        }
// cp1h := r.StoreCheckpoint()
// head := r.UpdateRef("heads/master", cp1h)
// tag1 := r.UpdateRef("tags/releaseDate", "message", u)

// Heavily inspired by Git.... this "write side" contains all the checkpointing
// metadata, etc required to know how a doain ev partition came into being
//
// It also has the nice side-effect that ev objects can be reused to save
// space (e.g lots of "accept_tos=1") events can be normalized
//
// I think that there should be a ref object that allows us to have branches
// and heads (since commits have parents it would allow us to checkpoint
// arbitrary things, and refer to "mainline" parent checkpoints, ala Git
// branching models)
//
// Because commits are "branchable" and affixes are bound to partitions
// there may be a race condition where you have a stale parent checkpoint
// and then fail to be able to commit, but Git solves this by implicitly
// making a branch, or having commits with two parents
