package main

import (
	"fmt"
	"testing"

	basictracer "github.com/opentracing/basictracer-go"
	"github.com/sirupsen/logrus"
)

var spanRecorder = basictracer.NewInMemoryRecorder()

func appFactory(t *testing.T) Application {
	r := Resolver{}

	// widgetsApp := &aggregates.WidgetsApp{}

	// r.Register(
	// 	"_",
	// 	func() (Aggregate, error) { return widgetsApp, nil },
	// 	map[string]func(aggregates.Aggregate) Command{
	// 		"AllowCreationOfNewIdentities": func(agg aggregates.Aggregate) Command {
	// 			return &AllowCreationOfNewIdentities{agg}
	// 		},
	// 	},
	// )

	// TODO: replace with noop tracer
	tracer := basictracer.New(spanRecorder)

	return &App{
		log:      logrus.StandardLogger(),
		tracer:   tracer,
		depot:    &MemoryRepository{},
		resolver: r.Resolve,
	}
}

func Test_App_ApplyRaiseErrOnNonRoutableCommand(t *testing.T) {
	t.Skip("not implemented yet")
}

func Test_App_InitializationWithConfigVars(t *testing.T) {

	var fixture []Cmd = []Cmd{
		{name: "AllowCreationOfNewIdentities"},
		// {name: "AllowCreationOfNewAuthorizations"},
		// {name: "_/AllowCreationOfNewAuthorizations"},
		// {name: "AllowBindingIdentitesToAuthorizations"},
		// {name: "USERS/123/SetPassword"},
		// {name: "users/123/SetUsername"},
	}
	a := appFactory(t)
	for _, cmd := range fixture {
		res, err := a.Apply(nil, "", cmd)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("apply res", cmd, "is", res)
	}
}
