package aggregates

import (
	"testing"

	test "github.com/retro-framework/go-retro/framework/test_helper"
)

func Test_Session_State_IsAnonymousByDefault(t *testing.T) {

	t.Skip("this test poses a legit question about how/where to set agg defaults (in factory, probably)")

	// Arrange
	assertBoolEql := test.H(t).BoolEql

	// Act
	session := Session{}

	// Assert
	assertBoolEql(session.HasIdentity, false)
}
