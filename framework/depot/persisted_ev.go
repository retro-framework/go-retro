package depot

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/framework/types"
)

type PersistedEv struct {
	time  time.Time
	name  string
	bytes []byte

	eventManifest types.EventManifest

	// The hash of the checkpoint which referred to the affix
	// from which this event was retrieved/unpacked.
	cpHash types.Hash
}

func (pEv PersistedEv) Time() time.Time {
	return pEv.time
}

func (pEv PersistedEv) Name() string {
	return pEv.name
}

func (pEv PersistedEv) Bytes() []byte {
	return pEv.bytes
}

func (pEv PersistedEv) CheckpointHash() types.Hash {
	return pEv.cpHash
}

func (pEv PersistedEv) Event() (types.Event, error) {
	var evFromManifest types.Event
	evFromManifest, err := pEv.eventManifest.ForName(pEv.Name())
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("can't retrieve event type from event manfest %#v", pEv.eventManifest))
	}
	err = json.Unmarshal(pEv.Bytes(), &evFromManifest)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("can't unmarshal json into restored event type %#v", pEv.eventManifest))
	}
	return evFromManifest, nil
}
