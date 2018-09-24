package memory

import "github.com/retro-framework/go-retro/framework/types"

func (r *RefStore) Reset(a, b, c, d bool) {
	r.r = make(map[string]types.Hash)
}
