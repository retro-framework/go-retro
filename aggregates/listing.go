package aggregates

import (
	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/framework/retro"
)

type Listing struct {
	NamedAggregate

	IsPublished          bool `json:"isPublic"`
	AuthorizedIdentities map[retro.PartitionName]string
}

func (agg *Listing) ReactTo(aev retro.Event) error {
	switch ev := aev.(type) {
	default:
		return errors.Errorf("Session aggregate doesn't know what to do with %s", ev)
	}
	return nil
}

func init() {
	Register("identity", &Identity{})
}
