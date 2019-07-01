package matcher

import "time"
import "sort"

// import "github.com/retro-framework/go-retro/framework/packing"

// Matchers always start at the tip of a given branch
// that means that we might have more than one ancestor
// at any point in the chain. We need to walk all the
// checkpoints and find the earliest history.
//
// This implementation takes a CheckpointHash and a reason
// along with a normalized timestamp and puts the checkpoints
// into an ordered array using a binary search. 
// 
// The structure itself is not concurrency safe, but it should be
// easy to make it concurrency safe by using channels to guard
// access to it.

type positionalResult struct {
	cpHash string
	t      time.Time
}

// Results typealias for []positionalResult with an insert
// method which will insert the entry and the correct location.
type Results []positionalResult

// Insert hash (retro.Hash) at time t. By searching the array
// to find a suitable place for it.
// 
// This always returns a new results, as the underlying heap
// storage for the slices should be reused anyway, and middling
// with pointers in here gets tedious as you can't *r[0] because
// it's not indexable.
func (r Results) Insert(t time.Time, h string) Results {
	var index = sort.Search(len(r), func(i int) bool { 
		return r[i].t.Equal(t) || r[i].t.After(t)
	})
	return append(r[:index], append([]positionalResult{positionalResult{h, t}}, r[index:]...)...)
}