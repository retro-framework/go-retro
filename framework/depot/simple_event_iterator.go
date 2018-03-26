package depot

import (
	"fmt"

	"github.com/retro-framework/go-retro/framework/types"
)

// SimpleEventIter emits events on a given partition
type SimpleEventIterator struct {
	pattern string
	c       chan types.PersistedEvent
}

func (s *SimpleEventIterator) Pattern() string {
	return s.pattern
}

func (s *SimpleEventIterator) Next() {
	fmt.Println("next called on simplepartition")
}

func (s *SimpleEventIterator) Events() (<-chan types.PersistedEvent, types.CancelFunc) {
	if s.c == nil {
		s.c = make(chan types.PersistedEvent)
	}
	return s.c, func() { close(s.c) }
}

func (s *SimpleEventIterator) pushCheckpoint(st *cpAffixStack) {

}
