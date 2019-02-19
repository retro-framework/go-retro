package packing

import "github.com/retro-framework/go-retro/framework/retro"

const (
	ObjectTypeAffix      retro.ObjectTypeName = "affix"
	ObjectTypeCheckpoint retro.ObjectTypeName = "checkpoint"
	ObjectTypeEvent      retro.ObjectTypeName = "event"

	ObjectTypeUnknown retro.ObjectTypeName = "unknown object type"
)

var KnownObjectTypes []retro.ObjectTypeName = []retro.ObjectTypeName{ObjectTypeAffix, ObjectTypeCheckpoint, ObjectTypeEvent}
