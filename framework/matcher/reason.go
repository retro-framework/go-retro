package matcher

// Reason is a simple enum so that queryable can keep
// track of what matched and why.
type Reason int

const (

	// ReasonNone indicates no match, implemented so that
	// callers can rely on having a result instead
	// of having to deal with nils.
	ReasonNone Reason = iota + 1

	// ReasonUnknown indicates no data available for a match
	// likely because a convenience API was used and not
	// enough context was available.
	ReasonUnknown

	// ReasonCheckpoint indicates that the matcher reported
	// a match as a checkpoint match.
	ReasonCheckpoint

	// ReasonAffix indicates that the matcher reported
	// a match as an affix match.
	ReasonAffix

	// ReasonEvent indicates that the matcher reported
	// a match as an event match. This is significant
	// because
	ReasonEvent
)
