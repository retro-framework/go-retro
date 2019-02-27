package retro

// Matcher is a generic matcher for searching
type Matcher interface {
	DoesMatch(interface{}) (bool, error)
}
