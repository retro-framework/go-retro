package repository

import (
	"context"

	"github.com/retro-framework/go-retro/framework/depot"
)

// TODO: make this respect the actual value that might come in a context
// TODO: move this to storage package?
func refFromCtx(ctx context.Context) string {
	return depot.DefaultBranchName
}
