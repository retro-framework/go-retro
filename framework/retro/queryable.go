package retro

import "context"

type MatcherResults interface{}

type Queryable interface {
	// For arbitrary queries of the storage space. MatcherResults
	// can be drained/filtered to extract Checkpoints, Affixes or
	// Events at various levels of granularity 
	Matching(context.Context, Matcher) (MatcherResults, error)
}
