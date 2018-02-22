// +build !redis

package redis

import (
	"fmt"
	"os"

	"github.com/retro-framework/go-retro/framework/types"
)

func NewDepot(evm types.EventManifest, nowFn types.NowFn) bool {
	fmt.Fprintf(os.Stderr, "error: attempted to use Redis depot, was built without tag `redis'.\n")
	os.Exit(1)
	return false
}
