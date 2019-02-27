package events

import (
	"encoding/json"
	"testing"

	"github.com/retro-framework/go-retro/framework/retro"
	test "github.com/retro-framework/go-retro/framework/test_helper"
)

type IdentityAggregateDouble struct {
	Name string `json:"urn"`
}

func (aid IdentityAggregateDouble) URN() retro.URN {
	return retro.URN(aid.Name)
}

func Test_AssociateIdentity(t *testing.T) {

	t.Run("marshalling", func(t *testing.T) {
		var aiEv = &AssociateIdentity{
			Identity: IdentityAggregateDouble{"identity/83a3097a45e74423dce29cdb"},
		}
		b, err := json.Marshal(aiEv)
		test.H(t).IsNil(err)
		test.H(t).StringEql(`{"identity":{"urn":"identity/83a3097a45e74423dce29cdb"}}`, string(b))
	})

	t.Run("unmarshalling", func(t *testing.T) {
		var (
			aiEv               = AssociateIdentity{}
			oldMarshalResult   = `{"identity":{"hasAvatar":false,"isPublic":false,"name":"identity/83a3097a45e74423dce29cdb"}}`
			idealMarshalResult = `{"identity":{"urn":"identity/83a3097a45e74423dce29cdb"}}`
		)

		err := json.Unmarshal([]byte(oldMarshalResult), &aiEv)
		test.H(t).IsNil(err)

		err = json.Unmarshal([]byte(idealMarshalResult), &aiEv)
		test.H(t).IsNil(err)

	})

}
