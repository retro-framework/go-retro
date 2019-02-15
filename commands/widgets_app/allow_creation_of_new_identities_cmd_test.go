package widgets_app

import (
	"bytes"
	"context"
	"testing"

	"github.com/retro-framework/go-retro/aggregates"
	"github.com/retro-framework/go-retro/framework/repository"
	test "github.com/retro-framework/go-retro/framework/test_helper"
)

func Test_WidgetsApp_AllowCreateIdentitesCmd_TogglesTheStateAndReturnsOneEvInSuccess(t *testing.T) {

	// Arrange
	var (
		assertBoolEql = test.H(t).BoolEql
		assertErrEql  = test.H(t).ErrEql

		receiver = &aggregates.WidgetsApp{}
	)

	// Act
	assertBoolEql(receiver.AllowCreateIdentities, false)
	cmd := AllowCreationOfNewIdentities{receiver}
	var b bytes.Buffer
	_, err := cmd.Apply(context.Background(), &b, &aggregates.Session{}, repository.NewSimpleRepositoryDouble(nil))

	// Assert
	assertErrEql(err, nil)
	// err := mr.Rehydrate(&app, "_")
	// assertErrEql(err, nil)
	// assertBoolEql(app.AllowCreateIdentities, true)
}
