package packing

import "github.com/retro-framework/go-retro/framework/types"

const (
	ObjectTypeAffix      types.ObjectTypeName = "affix"
	ObjectTypeCheckpoint types.ObjectTypeName = "checkpoint"
	ObjectTypeEvent      types.ObjectTypeName = "event"

	ObjectTypeUnknown types.ObjectTypeName = "unknown object type"
)

var KnownObjectTypes []types.ObjectTypeName = []types.ObjectTypeName{ObjectTypeAffix, ObjectTypeCheckpoint, ObjectTypeEvent}
