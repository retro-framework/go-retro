package retro

import "time"

// Clock allows dependency injection of a function returning
// the current time. Due to all the test code dealing with serialized
// data a clock with a predictable step is used extensively.
type Clock interface {
	Now() time.Time
}
