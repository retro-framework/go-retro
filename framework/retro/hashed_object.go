package retro

// HashedObject is a simple interface allowing more than one type
// of object to be hashed without knowing the type up-front.
// consumers can switch on the result of Type() and cast explicitly
// to one or the other type.
type HashedObject interface {
	Type() ObjectTypeName
	Contents() []byte
	Hash() Hash
}
