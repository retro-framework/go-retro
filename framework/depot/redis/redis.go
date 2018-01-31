// +build redis

package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis"
	opentracing "github.com/opentracing/opentracing-go"
	json "github.com/retro-framework/go-retro/framework/in-memory/freezer-json"
	"github.com/retro-framework/go-retro/framework/types"
)

type Error struct {
	Op  string
	Err error
}

func (e Error) Error() string {
	return fmt.Sprintf("redisdepot: op: %q err: %q", e.Op, e.Err)
}

type depot struct {
	client   *redis.Client
	evm      types.EventManifest
	nowFn    types.NowFn
	freezeFn func(types.Event) ([]byte, error)
}

func NewDepot(evm types.EventManifest, nowFn types.NowFn) *depot {
	// TODO: Check for stale locks ?
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	client.Ping()
	freezer := json.NewFreezer(evm)
	return &depot{client, evm, nowFn, freezer.Freeze}
}

func (d *depot) Claim(ctx context.Context, path string) bool {
	return d.lockPartition(ctx, path)
}

func (d *depot) Release(path string) {
	d.unlockPartition(path)
}

func (d *depot) Rehydrate(ctx context.Context, dest types.Aggregate, path string) error {

	spnRehydrate, ctx := opentracing.StartSpanFromContext(ctx, "depot/redis.Rehydrate")
	defer spnRehydrate.Finish()

	d.lockPartition(ctx, path)
	defer d.unlockPartition(path)

	// for _, ev := range d.aggEvs[path] {
	// 	spnReactToEv := opentracing.StartSpan("aggregate react to ev", opentracing.ChildOf(spnRehydrate.Context()))
	// 	spnReactToEv.LogKV("ev.object", ev)
	// 	err = dest.ReactTo(ev)
	// 	spnReactToEv.Finish()
	// 	if err != nil {
	// 		err := Error{"react-to", err}
	// 		spnRehydrate.LogKV("event", "error", "error.object", err)
	// 		return err
	// 	}
	// }

	return nil
}

func (d *depot) AppendEvs(path string, evs []types.Event) (int, error) {
	pipe := d.client.TxPipeline()
	for i, ev := range evs {
		b, err := d.freezeFn(ev)
		if err != nil {
			return i + 1, nil
		}
		pipe.RPush(path, string(b))
	}
	_, err := pipe.Exec()
	if err != nil {
		return 0, Error{"append-evs", err}
	}
	return len(evs), nil
}

func (d *depot) Exists(path string) bool {
	return d.client.Exists(path).Val() > int64(0)
}

func (d *depot) GetByDirname(ctx context.Context, path string) types.AggregateItterator {
	return nil
}

func (d *depot) lockPartition(ctx context.Context, path string) bool {
	deadline, ok := ctx.Deadline()
	if !ok {
		fmt.Fprintf(os.Stderr, "a context with no deadline was set to the ctx, a lock has been set in the redis depot with no ttl\n")
	}
	// TODO: check we *actually* lock don't just bounce off, to sim the behaviour
	// of sync.Mutex we should block until ctx is cancelled or lock is acquired
	return d.client.SetNX(d.lockKey(path), d.nowFn().Format(time.RFC3339Nano), deadline.Sub(time.Now())).Val()
}

func (d *depot) unlockPartition(name string) {
	d.client.Del(d.lockKey(name))
}

func (d *depot) lockKey(path string) string {
	return fmt.Sprintf("%s:lock", path)
}
