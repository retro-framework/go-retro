package aggregates

import (
	"testing"

	"github.com/retro-framework/go-retro/events"
	test "github.com/retro-framework/go-retro/framework/test_helper"
)

func Test_WidgetsApp_State_AllowCreationOfIdentities(t *testing.T) {
	// Arrange
	assertBoolEql := test.H(t).BoolEql
	assertErrEql := test.H(t).ErrEql
	mr := test.StateFixture(t, test.AggStateFixture("_", &events.AllowCreateIdentities{}))

	// Act
	app := WidgetsApp{}
	assertBoolEql(app.AllowCreateIdentities, false)

	// Assert
	err := mr.Rehydrate(&app, "_")
	assertErrEql(err, nil)
	assertBoolEql(app.AllowCreateIdentities, true)
}

func Test_WidgetsApp_State_DisableCreationOfIdentities(t *testing.T) {
	// Arrange
	assertBoolEql := test.H(t).BoolEql
	assertErrEql := test.H(t).ErrEql
	mr := test.StateFixture(t, test.AggStateFixture("_", &events.AllowCreateIdentities{},
		&events.DisableCreateIdentities{}))

	// Act
	app := WidgetsApp{}
	assertBoolEql(app.AllowCreateIdentities, false)

	// Assert
	err := mr.Rehydrate(&app, "_")
	assertErrEql(err, nil)
	assertBoolEql(app.AllowCreateIdentities, false)
}
