package json

import (
	"encoding/json"

	"github.com/retro-framework/go-retro/framework/types"
)

type envelope struct {
	Type    string      `json:"t"`
	Payload types.Event `json:"p"`
}

func NewFreezer(m types.EventManifest) *freezer {
	return &freezer{m}
}

type freezer struct {
	m types.EventManifest
}

func (f *freezer) Freeze(ev types.Event) ([]byte, error) {
	return json.Marshal(envelope{f.m.KeyFor(ev), ev})
}
