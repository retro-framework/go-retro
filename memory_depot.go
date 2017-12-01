package main

import (
	"github.com/leehambley/ls-cms/aggregates"
	"github.com/leehambley/ls-cms/events"
	"github.com/leehambley/ls-cms/storage"
	"github.com/pkg/errors"
)

type MemoryRepository struct {
	aggEvs map[string][]events.Event
}

// Rehydrate ideally receives a zero value object of the given aggregate
// type and then replays events on it to bring it into a known good state.
//
// At this level of abstraction we are not concerned with /commands/, this
// is purely reconstitution of pre-existing state ready to receive new commands
// by putting this aggregate in the firing line for new command handlers.
//
// Because I can't mutate `dest` here, as it's not passed as a pointer, we take
// whatever zero value we're given (from the upstream aggregate facory) which
// is registered with the resolver, and we return the modified one.
func (mr *MemoryRepository) Rehydrate(dest storage.Hydrated, path string) error {
	var err error
	for i, ev := range mr.aggEvs[path] {
		err = dest.ReactTo(ev)
		if err != nil {
			return errors.Errorf("error applying %#v to %s: %s (agg ev %d of %d)", ev, dest, err, i+1, len(mr.aggEvs[path]))
		}
	}
	return nil
}

func (mr *MemoryRepository) GetByDirname(path string) storage.HydratedItterator {
	return nil
}

// appendAggregateEvs appends the events to the history of the aggregate.
// It is *vital* that this is never called without the guarantees that the
// Command layer offers about the Aggregate's ability to accept these
// events, or react to them in a sane way.
func (mr *MemoryRepository) appendAggregateEvs(a aggregates.Aggregate, evs []events.Event) (int, int) {
	var urn = "urn-sentinel"
	// if _, ok := mr.aggregates[urn]; ok {
	mr.aggEvs[urn] = append(mr.aggEvs[urn], evs)
	return len(mr.aggEvs[urn]) - len(evs), len(mr.aggEvs[urn])
	// }
}
