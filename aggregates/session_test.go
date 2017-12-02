package aggregates

import (
	"testing"

	test "github.com/leehambley/ls-cms/framework/test_helper"
)

func Test_Session_State_IsAnonymousByDefault(t *testing.T) {

	t.Skip("this test poses a legit question about how/where to set agg defaults (in factory, probably)")

	// Arrange
	assertBoolEql := test.H(t).BoolEql

	// Act
	app := Session{}

	// Assert
	assertBoolEql(app.IsAnonymous, true) // TODO: ZeroValues ?
}
