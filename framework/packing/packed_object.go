package packing

import (
	"bytes"

	"github.com/retro-framework/go-retro/framework/types"
)

// Packed event represents a packed event in memory. The payload is a zlib
// deflated string of whatever the pack encoding scheme encodes to (json,
// msgpack, etc).
type po struct {
	payload []byte
	hash    types.Hash
}

func NewPackedObject(payloadStr string) types.HashedObject {
	return po{
		payload: []byte(payloadStr),
		hash:    hashStr(payloadStr),
	}
}

// Type returns a ObjectTypeName of either Affix, Checkpoint or Event
func (p po) Type() types.ObjectTypeName {
	parts := bytes.SplitN(p.payload, []byte(" "), 2)
	for _, kot := range KnownObjectTypes {
		if kot == types.ObjectTypeName(string(parts[0])) {
			return kot
		}
	}
	return ObjectTypeUnknown
}

func (p po) Contents() []byte {
	return p.payload
}

func (p po) Hash() types.Hash {
	return p.hash
}

type PackedEvent struct {
	types.HashedObject
}

func (pe PackedEvent) TypeName() types.ObjectTypeName {
	return ObjectTypeEvent
}

type PackedAffix struct {
	types.HashedObject
}

func (pe PackedAffix) TypeName() types.ObjectTypeName {
	return ObjectTypeAffix
}

type PackedCheckpoint struct {
	types.HashedObject
}

func (pc PackedCheckpoint) TypeName() types.ObjectTypeName {
	return ObjectTypeCheckpoint
}
