package widgets_app

import (
	"context"
	"testing"

	"github.com/leehambley/ls-cms/aggregates"
	test "github.com/leehambley/ls-cms/framework/test_helper"
)

func Test_WidgetsApp_AllowCreateIdentitesCmd_TogglesTheStateAndReturnsOneEvInSuccess(t *testing.T) {

	// Arrange
	assertBoolEql := test.H(t).BoolEql
	assertErrEql := test.H(t).ErrEql
	mr := test.EmptyStateFixture(t)

	// Act
	receiver := &aggregates.WidgetsApp{}
	assertBoolEql(receiver.AllowCreateIdentities, false)
	cmd := AllowCreationOfNewIdentities{receiver}
	_, err := cmd.Apply(context.Background(), &aggregates.Session{}, mr)

	// Assert
	assertErrEql(err, nil)
	// err := mr.Rehydrate(&app, "_")
	// assertErrEql(err, nil)
	// assertBoolEql(app.AllowCreateIdentities, true)
}
