//
// +build unit

package index

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/retro"
	"github.com/retro-framework/go-retro/framework/storage/memory"
	"github.com/retro-framework/go-retro/framework/test_helper"
)

type Predictable5sJumpClock struct {
	t     time.Time
	calls int
}

func (c *Predictable5sJumpClock) Now() time.Time {
	var next = c.t.Add(time.Duration((5 * c.calls)) * time.Second)
	c.calls = c.calls + 1
	return next
}

func Test_Simple(t *testing.T) {

	var jp = packing.NewJSONPacker()

	var clock = Predictable5sJumpClock{}

	// Checkpoints
	var (
		checkpointAlpha, _ = jp.PackCheckpoint(packing.Checkpoint{
			CommandDesc: []byte(`{"create":"author"}`),
			Fields: map[string]string{
				"session": "hello alpha",
				"date":    clock.Now().Format(time.RFC3339),
			},
		})

		checkpointOne, _ = jp.PackCheckpoint(packing.Checkpoint{
			CommandDesc: []byte(`{"create":"author"}`),
			Fields: map[string]string{
				"session": "hello world",
				"date":    clock.Now().Format(time.RFC3339),
			},
		})

		checkpointBeta, _ = jp.PackCheckpoint(packing.Checkpoint{
			CommandDesc: []byte(`{"draft":"article"}`),
			Fields: map[string]string{
				"session": "hello beta",
				"date":    clock.Now().Format(time.RFC3339),
			},
			ParentHashes: []retro.Hash{checkpointAlpha.Hash()},
		})

		checkpointTwo, _ = jp.PackCheckpoint(packing.Checkpoint{
			CommandDesc: []byte(`{"draft":"article"}`),
			Fields: map[string]string{
				"session": "hello world",
				"date":    clock.Now().Format(time.RFC3339),
			},
			ParentHashes: []retro.Hash{checkpointOne.Hash()},
		})

		checkpointThree, _ = jp.PackCheckpoint(packing.Checkpoint{
			CommandDesc: []byte(`{"update":"article"}`),
			Fields: map[string]string{
				"session": "hello world",
				"date":    clock.Now().Format(time.RFC3339),
			},
			ParentHashes: []retro.Hash{checkpointTwo.Hash()},
		})

		// Extend
		checkpointFourA, _ = jp.PackCheckpoint(packing.Checkpoint{
			CommandDesc: []byte(`{"update":"article"}`),
			Fields: map[string]string{
				"session": "hello world",
				"date":    clock.Now().Format(time.RFC3339),
			},
			ParentHashes: []retro.Hash{checkpointThree.Hash()},
		})

		checkpointFourB, _ = jp.PackCheckpoint(packing.Checkpoint{
			CommandDesc: []byte(`{"update":"article"}`),
			Fields: map[string]string{
				"session": "hello world",
				"date":    clock.Now().Format(time.RFC3339),
			},
			ParentHashes: []retro.Hash{checkpointThree.Hash(), checkpointFourA.Hash()},
		})
	)

	var (
		odb   = &memory.ObjectStore{}
		refdb = &memory.RefStore{}
	)

	odb.WritePacked(checkpointAlpha)
	odb.WritePacked(checkpointBeta)

	odb.WritePacked(checkpointOne)
	odb.WritePacked(checkpointTwo)
	odb.WritePacked(checkpointThree)
	odb.WritePacked(checkpointFourA)
	odb.WritePacked(checkpointFourB)
	refdb.Write("refs/heads/master", checkpointFourB.Hash())

	var checkpoints = map[string]retro.HashedObject{
		"one": checkpointOne,
		"two": checkpointTwo,
		"thr": checkpointThree,
		"foA": checkpointFourA,
		"foB": checkpointFourB,
		"alp": checkpointAlpha,
		"bet": checkpointBeta,
	}

	var H = test_helper.H(t)

	var expected = []retro.Hash{
		checkpointOne.Hash(),
		checkpointTwo.Hash(),
		checkpointThree.Hash(),
		checkpointFourA.Hash(),
		checkpointFourB.Hash(),
	}

	for id, checkpoint := range checkpoints {
		t.Log(id, "Checkpoint:", checkpoint.Hash().String())
	}

	//
	// 	(1)-->(2)-->(3)-->(4a)-->(4b)
	//              \----------^
	//
	// In in words, one, two three, and
	// four-a are linearly decended. four-b
	// is decended from three and four-a.
	//
	// NOTE: The implementation of Chronological
	// does not return until the index is built. Effectively
	// this has no impact because the channel will not yield
	// results until the index is built anyway, but this is
	// for that reason a blocking operation.
	t.Run("should emit checkpoints in the correct order", func(t *testing.T) {

		ctx, cancelFn := context.WithCancel(context.Background()) // TODO onbe from T?
		master := Chronological(ctx, odb, refdb)

		res, _ := readNFromCh(t, master, 5)
		H.StringEql(sprintHashSlice(res), sprintHashSlice(expected))
		cancelFn()

	})

}

func readNFromCh(t *testing.T, ch <-chan retro.Hash, n int) ([]retro.Hash, error) {
	var hashes []retro.Hash
	for i := 0; i < n; i++ {
		select {
		case <-time.After(1 * time.Second):
			return hashes, fmt.Errorf("deadline exceeded")
		case hash := <-ch:
			t.Log("received", hash)
			hashes = append(hashes, hash)
		}
	}
	return hashes, nil
}

func sprintHashSlice(h []retro.Hash) string {
	var str strings.Builder
	for i, r := range h {
		str.WriteString(fmt.Sprintf("%03d: %s\n", i, r.String()))
	}
	return str.String()
}
