package packing

import (
	"bytes"

	"github.com/retro-framework/go-retro/framework/retro"
)

// Packed event represents a packed event in memory. The payload is a zlib
// deflated string of whatever the pack encoding scheme encodes to (json,
// msgpack, etc).
type po struct {
	payload []byte
	hash    retro.Hash
}

func NewPackedObject(payloadStr string) retro.HashedObject {
	return po{
		payload: []byte(payloadStr),
		hash:    hashStr(payloadStr),
	}
}

// Type returns a ObjectTypeName of either Affix, Checkpoint or Event
func (p po) Type() retro.ObjectTypeName {
	parts := bytes.SplitN(p.payload, []byte(" "), 2)
	for _, kot := range KnownObjectTypes {
		if kot == retro.ObjectTypeName(string(parts[0])) {
			return kot
		}
	}
	return ObjectTypeUnknown
}

func (p po) Contents() []byte {
	return p.payload
}

func (p po) Hash() retro.Hash {
	return p.hash
}

type PackedEvent struct {
	retro.HashedObject
}

func (pe PackedEvent) TypeName() retro.ObjectTypeName {
	return ObjectTypeEvent
}

type PackedAffix struct {
	retro.HashedObject
}

func (pe PackedAffix) TypeName() retro.ObjectTypeName {
	return ObjectTypeAffix
}

type PackedCheckpoint struct {
	retro.HashedObject
}

func (pc PackedCheckpoint) TypeName() retro.ObjectTypeName {
	return ObjectTypeCheckpoint
}
