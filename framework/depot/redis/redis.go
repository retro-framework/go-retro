// +build redis

package redis

import (
	"github.com/go-redis/redis"
	json "github.com/retro-framework/go-retro/framework/in-memory/freezer-json"
	"github.com/retro-framework/go-retro/framework/types"
)

func NewDepot(evm types.EventManifest, nowFn types.NowFn) *depot {
	// TODO: Check for stale locks ?
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	client.Ping()
	freezer := json.NewFreezer(evm)
	return &depot{client, evm, nowFn, freezer.Freeze}
}
