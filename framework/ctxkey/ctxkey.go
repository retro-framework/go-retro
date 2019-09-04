package ctxkey

import "context"

type contextKey string

func (c contextKey) String() string {
	return "retro ref " + string(c)
}

var (
	contextKeyRef = contextKey("ref")
)

// Ref gets the ref name from the context. If none
// is present then a default value of "refs/heads/master" is returned
// which is indicated to callers by setting the 2nd return parameter
// true.
func Ref(ctx context.Context) string {
	ref, ok := ctx.Value(contextKeyRef).(string)
	if ref == "" || !ok {
		// TODO: should be depot.DefaultBranchName
		return "refs/heads/master"
	}
	return ref
}
