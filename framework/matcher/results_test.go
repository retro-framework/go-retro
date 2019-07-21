// +build unit

package matcher

import (
	"testing"
)

// Test_Results the Results type takes a timestamp
// and a checkpoint hash (as a string, because of avoiding
// import cycles). The existing set of things is searched
// using a binary search (probably) and a new set of results
// is returned with the checkpoints ordered by the time
// at which they went into the database.
func Test_Results(t *testing.T) {

	// var toc = TimeOrderedCheckpoints{}

	// toc = toc.Insert(time.Now().Add(-5*time.Minute), "one")
	// toc = toc.Insert(time.Now().Add(-15*time.Minute), "two")
	// toc = toc.Insert(time.Now().Add(-45*time.Minute), "three")

	// var expectedResults = []string{"three", "two", "one"}
	// for i, eR := range expectedResults {
	// 	if toc[i].cpHash != eR {
	// 		t.Fatalf("Expecte %q at %d, found %q\n", eR, i, toc[i].cpHash)
	// 	}
	// }

}



1. 5 things in the index
2. consumer starts reading index
3. index is lazy evaluated first time
4. immediately two documents added to index
5. consumer receives 7 items
6. new consumer immediately receives 7 items
7. new documented added to index
8. both consumers receive 1 new item