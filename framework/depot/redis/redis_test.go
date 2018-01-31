// +build redis,integration

package redis

import "testing"

func Test_RedisDepot(t *testing.T) {
	NewDepot()
}
