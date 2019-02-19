package retro

import "io"

// DumpableDepot is an optional interface which implements
// a single method which dumps the contents as preformatted
// text to facilitate easy debugging. It is mostly used in
// integration tests.
type DumpableDepot interface {
	DumpAll(w io.Writer) string
}
