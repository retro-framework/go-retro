package memory

import (
	"errors"

	"github.com/retro-framework/go-retro/framework/packing"
)

var (
	ErrUnknownRef         = errors.New("ref name unknown")
	ErrUnknownSymbolicRef = errors.New("symbolic ref name unknown")
)

// RefStore is used for storing references, references such as
// refs/heads/master (branch) or HEAD (symbolic) or refs/wurtzel/booger for
// arbitrary checkpoints.
type RefStore struct {
	r map[string]packing.Hash
	s map[string]string
}

// Write ref returns a boolean indicating whether the ref was changed
// or not, and errors incase of malformation, and misc problems.
func (r *RefStore) Write(name string, newRef packing.Hash) (bool, error) {
	if r.r == nil {
		r.r = make(map[string]packing.Hash)
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

func (r *RefStore) Retrieve(name string) (*packing.Hash, error) {
	if existingRef, exists := r.r[name]; exists {
		return &existingRef, nil
	}
	return nil, ErrUnknownRef
}

func (r *RefStore) RetrieveSymbolic(name string) (string, error) {
	if existingRef, exists := r.s[name]; exists {
		return existingRef, nil
	}
	return "", ErrUnknownSymbolicRef
}
