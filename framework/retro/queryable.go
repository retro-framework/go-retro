package retro

import "context"

type Queryable interface {
	Matching(context.Context, Matcher) (interface{}, error) // TODO: nail this down a bit
}
