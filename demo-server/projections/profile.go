package projections

import (
	"github.com/go-redis/redis"
)

type Profile struct {
	Name    string
	Friends []Profile
}

type Profiles struct {
	redis redis.Client
}

func (p Profiles) Get(name string) Profile {
	return Profile{
		Name: "Lee H",
		Friends: []Profile{
			Profile{Name: "Jack A"},
			Profile{Name: "John B"},
			Profile{Name: "Jessie C"},
		},
	}
}
