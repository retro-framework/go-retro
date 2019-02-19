package retro

// NamedAggregate is used by the Engine to stealthily pass
// names through to the command, where appropriate. The
// interface should never be used in a Command.
//
// For known, loaded-from-storage aggregates the name can
// be set in the engine and survive the roundtrip to the
// command and be read back in the Engine after Apply has
// returned. The CommandResult map[retro.Aggregage][]retro.Event
// then allows the use of the interface via upgrade on
// the keys of the map to get the names to use when persisting
// the events.
type ExistingAggregate interface {
	Name() PartitionName
}
