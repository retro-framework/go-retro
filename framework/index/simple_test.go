// +build unit

package index

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/retro"
	"github.com/retro-framework/go-retro/framework/storage/memory"
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
		checkpointOne, _ = jp.PackCheckpoint(packing.Checkpoint{
			CommandDesc: []byte(`{"create":"author"}`),
			Fields: map[string]string{
				"session": "hello world",
				"date":    clock.Now().Format(time.RFC3339),
			},
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

	odb.WritePacked(checkpointOne)
	odb.WritePacked(checkpointTwo)
	odb.WritePacked(checkpointThree)
	odb.WritePacked(checkpointFourA)
	odb.WritePacked(checkpointFourB)
	refdb.Write("refs/heads/master", checkpointThree.Hash())
	// refdb.Write("refs/heads/alter", checkpointFourA.Hash())

	t.Run("should emit things in order", func(t *testing.T) {

		fmt.Println("1 ", checkpointOne.Hash().String())
		fmt.Println("2 ", checkpointTwo.Hash().String())
		fmt.Println("3 ", checkpointThree.Hash().String())
		fmt.Println("4a", checkpointFourA.Hash().String())
		fmt.Println("4b", checkpointFourB.Hash().String())

		var tooSlow = time.AfterFunc(5*time.Millisecond, func() {
			t.Fatal("test too too long to run")
		})

		master, masterCancelFn := ChronologicalForBranch("refs/heads/master", odb, refdb)
		defer masterCancelFn()

		if master.Next(context.TODO()).String() != checkpointOne.Hash().String() {
			t.Fatalf("expected first hash to be %q\n", checkpointOne.Hash())
		}
		if master.Next(context.TODO()).String() != checkpointTwo.Hash().String() {
			t.Fatalf("expected first hash to be %q\n", checkpointTwo.Hash())
		}
		if master.Next(context.TODO()).String() != checkpointThree.Hash().String() {
			t.Fatalf("expected first hash to be %q\n", checkpointThree.Hash())
		}

		go func(){
			refdb.Write("refs/heads/master", checkpointFourA.Hash())
			}()

		if master.Next(context.TODO()).String() != checkpointFourA.Hash().String() {
			t.Fatalf("expected first hash to be %q\n", checkpointFourA.Hash())
		}

		// if master.Next(context.TODO()).String() != checkpointFourB.Hash().String() {
		// 	t.Fatalf("expected first hash to be %q\n", checkpointFourB.Hash())
		// }

		tooSlow.Stop()

	})

}
