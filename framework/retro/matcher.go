package retro

import "github.com/retro-framework/go-retro/framework/matcher"

// Matcher is a generic matcher for searching
//
// TODO: Should we allow an interface upgrade for
// matchers, so that they can maybe signal that
// they are done matching to allow the queryable
// to skip doing a lot of IO on affix/events if the
// matcher is only interested in checkpoints?
// queryable could then cancel other io operations
// on those other objects if the matcher doesn't
// want to do anymore matching?
type Matcher interface {
	DoesMatch(interface{}) (matcher.Result, error)
}
