package memory

import (
	"fmt"
	"sync"
	"time"

	"github.com/retro-framework/go-retro/framework/retro"
	"github.com/retro-framework/go-retro/framework/storage"
)

// RefStore is used for storing references, references such as
// refs/heads/master (branch) or HEAD (symbolic) or refs/wurtzel/booger for
// arbitrary checkpoints.
type RefStore struct {
	r map[string]retro.Hash
	s map[string]string

	m []chan<- retro.RefMove

	sync.RWMutex
}

func (r *RefStore) Ls() (map[string]retro.Hash, error) {
	r.RLock()
	defer r.RUnlock()
	res := make(map[string]retro.Hash)
	for k, v := range r.r {
		res[k] = v
	}
	return res, nil
}

func (r *RefStore) NotifyOn(ch chan<- retro.RefMove) {
	r.Lock()
	defer r.Unlock()
	r.m = append(r.m, ch)
}

// Write ref returns a boolean indicating whether the ref was changed
// or not, and errors incase of malformation, and misc problems.
func (r *RefStore) Write(name string, newRef retro.Hash) (bool, error) {
	r.Lock()
	defer r.Unlock()

	var refMove retro.RefMove
	if r.r == nil {
		r.r = make(map[string]retro.Hash)
	}
	if existingRef, exists := r.r[name]; exists {
		if newRef.String() == existingRef.String() {
			return false, nil
		} else {
			refMove = retro.RefMove{
				Old:  existingRef,
				New:  newRef,
				Name: name,
			}
		}
	}
	refMove = retro.RefMove{
		Old:  nil,
		New:  newRef,
		Name: name,
	}
	r.r[name] = newRef
	for _, ch := range r.m {
		select {
		case ch <- refMove:
		case <-time.After(1 * time.Second):
			fmt.Println("timed out waiting to inform a subscriber about a ref move")
		}
	}
	return true, nil
}

func (r *RefStore) WriteSymbolic(name string, ref string) (bool, error) {
	if r.s == nil {
		r.s = make(map[string]string)
	}
	if existingRef, exists := r.s[name]; exists {
		if existingRef == ref {
			return false, nil
		}
	}
	r.s[name] = ref
	return true, nil
}

func (r *RefStore) Retrieve(name string) (retro.Hash, error) {
	if existingRef, exists := r.r[name]; exists {
		return existingRef, nil
	}
	return nil, storage.ErrUnknownRef
}

func (r *RefStore) RetrieveSymbolic(name string) (string, error) {
	if existingRef, exists := r.s[name]; exists {
		return existingRef, nil
	}
	return "", storage.ErrUnknownSymbolicRef
}
