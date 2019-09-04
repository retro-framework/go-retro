package index

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/retro-framework/go-retro/framework/retro"
)

// tocCP is a Table Of Contents, Checkpoint. The structure
// represents a linearized view of the merkle tree.
type tocCP struct {
	h retro.Hash
	t time.Time
}

type tocCPs []tocCP

// Insert inserts the tocCP at the correct position
// after binary searching for the appropriate place
func (r tocCPs) Insert(t tocCP) tocCPs {
	// early return if we know this checkpoint already
	for _, myT := range r {
		if bytes.Equal(t.h.Bytes(), myT.h.Bytes()) {
			return r
		}
	}
	var index = sort.Search(len(r), func(i int) bool {
		return r[i].t.Equal(t.t) || r[i].t.After(t.t)
	})
	return append(r[:index], append([]tocCP{t}, r[index:]...)...)
}

func (t tocCPs) String() string {
	var str strings.Builder
	str.WriteString("======================================\n")
	for i, tocCP := range t {
		var s = fmt.Sprintf(
			"> \t%s %s (%03d/%03d)\n",
			tocCP.t.Format(time.RFC3339),
			tocCP.h.String(),
			i,
			len(t),
		)
		str.WriteString(s)
	}
	return str.String()
}
