package retro

// CommandResult is a type alias for map[string][]Event
// to make the function signatures expressive. The resulting
// map should ideally contain
type CommandResult map[Aggregate][]Event
