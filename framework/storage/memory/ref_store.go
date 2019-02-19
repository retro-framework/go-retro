package memory

import (
	"github.com/retro-framework/go-retro/framework/storage"
	"github.com/retro-framework/go-retro/framework/retro"
)

// RefStore is used for storing references, references such as
// refs/heads/master (branch) or HEAD (symbolic) or refs/wurtzel/booger for
// arbitrary checkpoints.
type RefStore struct {
	r map[string]retro.Hash
	s map[string]string
}

func (r *RefStore) Ls() (map[string]retro.Hash, error) {
	return r.r, nil
}

// Write ref returns a boolean indicating whether the ref was changed
// or not, and errors incase of malformation, and misc problems.
func (r *RefStore) Write(name string, newRef retro.Hash) (bool, error) {
	if r.r == nil {
		r.r = make(map[string]retro.Hash)
	}
	if existingRef, exists := r.r[name]; exists {
		if newRef.String() == existingRef.String() {
			return false, nil
		}
	}
	r.r[name] = newRef
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
