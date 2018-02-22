// +build redis

package redis

import (
	"github.com/go-redis/redis"
	"github.com/retro-framework/go-retro/framework/types"
)

func NewDepot(evm types.EventManifest, nowFn types.NowFn) bool {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	client.Ping()
	return false
}
