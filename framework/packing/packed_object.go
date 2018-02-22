package packing

// Packed event represents a packed event in memory. The payload is a zlib
// deflated string of whatever the pack encoding scheme encodes to (json,
// msgpack, etc).
type PackedObject struct {
	payload []byte
	hash    Hash
}

func (po *PackedObject) Contents() []byte { return po.payload }
func (po *PackedObject) Hash() Hash       { return po.hash }

type PackedEvent struct{ PackedObject }

func (pe *PackedEvent) TypeName() ObjectTypeName { return ObjectTypeEvent }

type PackedAffix struct{ PackedObject }

func (pe *PackedAffix) TypeName() ObjectTypeName { return ObjectTypeAffix }

type PackedCheckpoint struct{ PackedObject }

func (pe *PackedCheckpoint) TypeName() ObjectTypeName { return ObjectTypeCheckpoint }
