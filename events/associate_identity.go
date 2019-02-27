package events

import (
	"encoding/json"

	"github.com/retro-framework/go-retro/framework/retro"
)

type AssociateIdentity struct {
	Identity retro.URNAble `json:"identity"`
}

func (ai *AssociateIdentity) UnmarshalJSON(b []byte) error {

	var tmp = struct {
		Identity struct {
			Name string `json:"name"`
			URN  string `json:"urn"`
		} `json:"identity"`
	}{}

	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	if tmp.Identity.Name != "" {
		ai.Identity = WeakURNReference{retro.URN(tmp.Identity.Name)}
	}

	if tmp.Identity.URN != "" {
		ai.Identity = WeakURNReference{retro.URN(tmp.Identity.URN)}
	}

	return nil
}

func init() {
	Register(&AssociateIdentity{})
}
