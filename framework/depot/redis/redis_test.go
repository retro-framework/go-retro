// +build redis,integration

package redis

import (
	"testing"
	"time"

	"github.com/retro-framework/go-retro/events"
)

func Test_RedisDepot(t *testing.T) {
	NewDepot(events.DefaultManifest, time.Now)
}
