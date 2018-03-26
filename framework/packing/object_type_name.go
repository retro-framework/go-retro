package packing

// ObjectTypeName represents the
type ObjectTypeName string

const (
	ObjectTypeAffix      ObjectTypeName = "affix"
	ObjectTypeCheckpoint ObjectTypeName = "checkpoint"
	ObjectTypeEvent      ObjectTypeName = "event"

	ObjectTypeUnknown ObjectTypeName = "unknown object type"
)

var KnownObjectTypes []ObjectTypeName = []ObjectTypeName{ObjectTypeAffix, ObjectTypeCheckpoint, ObjectTypeEvent}
