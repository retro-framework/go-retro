package retro

// RefMove represents a head pointer movement
// it contains the old and new hashes. A boolean
// is set indicating whether this is a FF move
// or not.
type RefMove struct {
	Old Hash
	New Hash

	Name string
}
