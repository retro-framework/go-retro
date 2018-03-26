package packing

import "testing"

func Test_HashStrToHash(t *testing.T) {

	h1 := hashStr("hello world")
	h2 := HashStrToHash(h1.String())

	if h1.String() != h2.String() {
		t.Fatalf("HashStrToHash failed %s was not equal to %s", h1.String(), h2.String())
	}

}
