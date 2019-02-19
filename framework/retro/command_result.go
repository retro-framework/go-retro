package retro

import (
	"encoding/json"
)

// CommandResult is a type alias for map[string][]Event
// to make the function signatures expressive.
type CommandResult map[Aggregate][]Event

// MarshalJSON is required because without the method
// override
func (cr CommandResult) MarshalJSON() ([]byte, error) {
	var t = make(map[PartitionName][]Event)
	for k, v := range cr {
		t[k.Name()] = v
	}
	return json.Marshal(t)
}
