package main

import (
	"testing"

	basictracer "github.com/opentracing/basictracer-go"
)

var spanRecorder = basictracer.NewInMemoryRecorder()

func Test_App_ApplyRaiseErrOnNonRoutableCommand(t *testing.T) {
	t.Skip("not implemented yet")
}

func Test_App_InitializationWithConfigVars(t *testing.T) {
	t.Skip("not implemented yet")
}
