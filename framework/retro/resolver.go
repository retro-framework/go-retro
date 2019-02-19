package retro

import "context"

// Resolver takes a []byte and returns a callable command function
// the resolver is used bt the Engine to take serialized client
// input and map it to a registered command by name. The command
// func returned will usually be a function on a struct type
// which the resolver will instantiate and prepare for execution.
type Resolver interface {
	Resolve(context.Context, Repository, []byte) (Command, error)
}
