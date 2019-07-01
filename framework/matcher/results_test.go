// +build unit

package matcher

import (
	"testing"
	"time"
)

// Test_Results the Results type takes a timestamp
// and a checkpoint hash (as a string, because of avoiding
// import cycles). The existing set of things is searched
// using a binary search (probably) and a new set of results
// is returned with the checkpoints ordered by the time
// at which they went into the database.
func Test_Results(t *testing.T) {

	var r = Results{}

	r = r.Insert(time.Now().Add(-5 * time.Minute),  "one")
	r = r.Insert(time.Now().Add(-15 * time.Minute), "two")
	r = r.Insert(time.Now().Add(-45 * time.Minute), "three")

	var expectedResults = []string{"three", "two", "one"}
	for i, eR := range expectedResults {
		if r[i].cpHash != eR {
			t.Fatalf("Expecte %q at %d, found %q\n", eR, i, r[i].cpHash)
		}
	}

}
