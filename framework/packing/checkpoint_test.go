package packing

import "testing"
import "github.com/retro-framework/go-retro/framework/test_helper"

func Test_Checkpoint_HasErrors(t *testing.T) {

	var h = test_helper.H(t)
	var _, errs = Checkpoint{}.HasErrors()
	t.Fatal(h.MustSerilizeYAML(errs))

}
