package retro

import "context"

// ResolveFn does the heavy lifting on the resolution. The Resolver
// interface is clumbsy for use in tests and the ResolveFn allows
// a simple anonymous drop-in in tests which can resolve a stub/double
// without lots of boilerplate code.
type ResolveFn func(context.Context, Repo, []byte) (Command, error)
