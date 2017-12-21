package memory

import (
	"context"
	"fmt"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/retro-framework/go-retro/framework/types"
)

type Error struct {
	Op  string
	Err error
}

func (e Error) Error() string {
	return fmt.Sprintf("memorydepot: op: %q err: %q", e.Op, e.Err)
}

type Depot struct {
	aggEvs map[string][]types.Event
}

func NewDepot(state map[string][]types.Event) *Depot {
	return &Depot{aggEvs: state}
}

func NewEmptyDepot() *Depot {
	return &Depot{aggEvs: map[string][]types.Event{}}
}

// Rehydrate ideally receives a zero value object of the given aggregate
// type and then replays events on it to bring it into a known good state.
//
// At this level of abstraction we are not concerned with /commands/, this
// is purely reconstitution of pre-existing state ready to receive new commands
// by putting this aggregate in the firing line for new command handlers.
//
// TODO: is this doc bogus? Because I can't mutate `dest` here, as it's not
// passed as a pointer, we take whatever zero value we're given (from the
// upstream aggregate facory) which is registered with the resolver, and
// we return the modified one.
func (d *Depot) Rehydrate(ctx context.Context, dest types.Aggregate, path string) error {

	spnRehydrate, ctx := opentracing.StartSpanFromContext(ctx, "memorydepot.Rehydrate")
	defer spnRehydrate.Finish()

	var err error
	for _, ev := range d.aggEvs[path] {
		spnReactToEv := opentracing.StartSpan("aggregate react to ev", opentracing.ChildOf(spnRehydrate.Context()))
		spnReactToEv.LogKV("ev.object", ev)
		err = dest.ReactTo(ev)
		spnReactToEv.Finish()
		if err != nil {
			err := Error{"react-to", err}
			spnRehydrate.LogKV("event", "error", "error.object", err)
			return err
		}
	}

	return nil
}

func (d *Depot) GetByDirname(ctx context.Context, path string) types.AggregateItterator {
	return nil
}

// appendAggregateEvs appends the events to the history of the aggregate.
// It is *vital* that this is never called without the guarantees that the
// Command layer offers about the Aggregate's ability to accept these
// events, or react to them in a sane way.
func (d *Depot) appendAggregateEvs(a types.Aggregate, evs []types.Event) (int, int) {
	var urn = "urn-sentinel"
	d.aggEvs[urn] = append(d.aggEvs[urn], evs)
	return len(d.aggEvs[urn]) - len(evs), len(d.aggEvs[urn])
}
