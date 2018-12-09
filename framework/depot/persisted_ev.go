package depot

import (
	"encoding/json"
	"fmt"
	"reflect"
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
	evFromManifest, err := pEv.eventManifest.ForName(pEv.Name())
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("can't retrieve event type from event manfest %#v", pEv.eventManifest))
	}
	err = json.Unmarshal(pEv.Bytes(), &evFromManifest)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("can't unmarshal json into restored event type %#v", pEv.eventManifest))
	}

	// ForName returns a pointer to a types.Event because of the way the reflection
	// works, this unwraps it else raises an error. This unwrapping could feasibly
	// be moved into the ForName function. For Aggregates the reasoning is different
	// so we can afford to return a pointer (aggregates are modified as they are re-hydrated
	// and thus need to be mutable as they are passed around, events are plain ol'
	// value objects)
	if reflect.TypeOf(evFromManifest).Kind() == reflect.Ptr {
		return reflect.ValueOf(evFromManifest).Elem().Interface(), nil
	}

	return evFromManifest, errors.Wrap(err, fmt.Sprintf("expected to get a pointer back from eventManifest.ForName()"))
}
