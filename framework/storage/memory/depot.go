package memory

import (
	"context"
	"fmt"
	"reflect"
	"sync"

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
	sync.RWMutex
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
func (d *Depot) Rehydrate(ctx context.Context, dest types.Aggregate, path string) error {

	spnRehydrate, ctx := opentracing.StartSpanFromContext(ctx, "memorydepot.Rehydrate")
	defer spnRehydrate.Finish()

	var err error

	d.RLock()
	defer d.RUnlock()

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

func (d *Depot) Claim(ctx context.Context, path string) bool {
	return true
}

func (d *Depot) Release(path string) {
	return
}

func (d *Depot) AppendEvs(path string, evs []types.Event) (int, error) {
	d.Lock()
	d.aggEvs[path] = append(d.aggEvs[path], evs)
	d.Unlock()
	return len(evs), nil
}

func (d *Depot) Exists(path string) bool {
	d.RLock()
	_, exists := d.aggEvs[path]
	d.RUnlock()
	return exists
}

func toType(t types.Aggregate) reflect.Type {
	var v = reflect.ValueOf(t)
	if reflect.Ptr == v.Kind() || reflect.Interface == v.Kind() {
		v = v.Elem()
	}
	return v.Type()
}
