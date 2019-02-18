package repository

import (
	"testing"

	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/storage/memory"
)

type dummyAggregate struct{}

func Test_simpleNew(t *testing.T) {
	t.Skip("no tests yet, repository basically defers to two other tested things anyway")
	var (
		objdb      = &memory.ObjectStore{}
		refdb      = &memory.RefStore{}
		evM        = events.NewManifest()
		repository = NewSimpleRepository(objdb, refdb, evM)
	)
	_ = repository
}
