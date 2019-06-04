package retro

import "context"

type MatcherResults interface{}

type Queryable interface {
	Matching(context.Context, Matcher) (MatcherResults, error)
}
