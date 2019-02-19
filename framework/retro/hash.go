package retro

// Hash is a minimal interface for the framework
// as a whole. It can broadly be considered as
// a type alias for string as most of the storage
// and "downstream" (to clients) code works with
// serialized data. There is a concrete implementation
// of hash which offers more methods but String()
// is lingua franca.
type Hash interface {
	String() string
	Bytes() []byte
}
