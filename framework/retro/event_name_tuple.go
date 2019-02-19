package retro

// EventNameTuple is used exclusively in the tests for constructing
// a test fixture. It could be moved into conditional compilation or
// defined solely in the tests potentially.
type EventNameTuple struct {
	Name  string
	Event Event
}
