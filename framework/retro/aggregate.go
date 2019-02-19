package retro

// Aggregate interface is simple, it need only ReactTo
// events. This Aggregate interface is only relevant to the
// write side of the application.
//
// Name is used by the Command objects to get a name for
// aggregates. If aggregates are anonymous then they may
// also receive SetName. Renaming aggregates is not supported
type Aggregate interface {
	ReactTo(Event) error

	// TODO: pull these out into a separate interface
	// as below so that we can keep this reserved
	// for the Engine internals.
	Name() PartitionName
	SetName(PartitionName) error
}
