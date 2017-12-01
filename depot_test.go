package main

import (
	"testing"

	"github.com/leehambley/ls-cms/aggregates"
	"github.com/leehambley/ls-cms/events"
)

func eventFixture() []events.Event {
	return []events.Event{
		&events.AllowCreateIdentities{},
		&events.SetAvatar{
			"image/gif",
			[]byte("DEADBEEF"),
		},
		&events.SetDisplayName{"mr. meeseeks"},
	}
}

func Test_Depot_CommandStorage(t *testing.T) {

}

func Test_Depot_CommandStorageRoundtrip(t *testing.T) {

}

func Test_Depot_EventStorage(t *testing.T) {

	mr := MemoryRepository{
		map[string][]events.Event{"_": eventFixture()},
	}
	app := aggregates.WidgetsApp{}
	err := mr.Rehydrate(&app, "_")
	if err != nil {
		t.Fatalf("error rehydrating app: %s", err)
	}

	t.Logf("App: %s", app)
	t.Logf("Event Manifest: %#v", events.Manifest)
	t.Logf("Aggregate Manifest: %#v", aggregates.Manifest)
}

func Test_Depot_EventStorageRoundtrip(t *testing.T) {

}
