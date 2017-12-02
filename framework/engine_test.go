package main

import (
	"testing"

	basictracer "github.com/opentracing/basictracer-go"
)

var spanRecorder = basictracer.NewInMemoryRecorder()

// func engineFactory(t *testing.T) types.StateEngine {
// 	r := resolver.New()
//
// 	// TODO: replace with noop tracer
// 	tracer := basictracer.New(spanRecorder)
//
// 	return &Engine{
// 		log:      logrus.StandardLogger(),
// 		tracer:   tracer,
// 		depot:    &memory.NewEmptyDepot(),
// 		resolver: r.Resolve,
// 	}
// }

func Test_App_ApplyRaiseErrOnNonRoutableCommand(t *testing.T) {
	t.Skip("not implemented yet")
}

func Test_App_InitializationWithConfigVars(t *testing.T) {

	// assertErrEql := h(t).ErrEql
	//
	// cmdFixture = ``
	//
	// e := engineFactory(t)
	// for _, cmd := range fixture {
	// 	res, err := e.Apply(nil, "", cmd)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// }
}
