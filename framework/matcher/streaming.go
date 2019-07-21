package matcher

// StreamingMatcher is the normal API for a matcher
// non streaming variants can be built on top with
// an escape hatch to stop upon seeing a specicific
// version. A StreamingMatcher needs to know how to
// behave on a head pointer move.
type StreamingMatcher struct {
}

// A StreamingMatcher takes a branch (ref) and one
// or more matchers.

// A StreamingMatcher looks up the chronology index
// for a given branch (ref) to get a linearized view
// of the history to date

// The chronology index may be modelled as a sorted
// array internally, however we will need to read
// it as a channel.

// The chronology index need not expose the underlying
// timestamps, but must give us the checkpoint hashes.

// Chronology indexes are maintained by the underlying
// storage when moving the head pointer. In simple
// cases this should be inexpensive as it's a simple apend
// in other cases it may be more expensive

// In any case, the StreamingMatcher should normally operate
// in a way that also receives new updates such as when
// subscribing to MoveHeadPointer changes in the depot.
