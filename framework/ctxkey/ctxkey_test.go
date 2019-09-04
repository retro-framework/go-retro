package ctxkey

import (
	"context"
	"testing"
)

func Test_ctxkey_Ref(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		res := Ref(context.Background())
		if res != "refs/heads/master" {
			t.Fatal("expected default value, got", res)
		}
	})
	t.Run("override", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), contextKeyRef, "refs/heads/other")
		res := Ref(ctx)
		if res != "refs/heads/other" {
			t.Fatal("expected overridden value, got", res)
		}
	})
}
