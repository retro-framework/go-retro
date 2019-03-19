package aggregates

import (
	"testing"

	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/retro"
	test "github.com/retro-framework/go-retro/framework/test_helper"
)

func Test_WidgetsApp_State_AllowCreationOfIdentities(t *testing.T) {
	// Arrange
	var (
		assertBoolEql = test.H(t).BoolEql
		assertErrEql  = test.H(t).ErrEql

		app     = &WidgetsApp{}
		fixture = retro.EventFixture{
			app: []retro.Event{
				&events.AllowCreateIdentities{},
			},
		}
	)

	// Act
	assertBoolEql(app.AllowCreateIdentities, false)

	// Assert
	err := test.H(t).Rehydrate(fixture, app)
	assertErrEql(err, nil)
	assertBoolEql(app.AllowCreateIdentities, true)
}

func Test_WidgetsApp_State_DisableCreationOfIdentities(t *testing.T) {
	// Arrange
	var (
		assertBoolEql = test.H(t).BoolEql
		assertErrEql  = test.H(t).ErrEql

		app     = &WidgetsApp{}
		fixture = retro.EventFixture{
			app: []retro.Event{
				&events.AllowCreateIdentities{},
				&events.DisableCreateIdentities{},
			},
		}
	)

	// Act
	assertBoolEql(app.AllowCreateIdentities, false)

	// Assert
	err := test.H(t).Rehydrate(fixture, app)
	assertErrEql(err, nil)
	assertBoolEql(app.AllowCreateIdentities, false)
}
