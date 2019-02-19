package retro

// SessionID is a type alias for string to convey the meaning that a real
// session ID is required and not any (maybe empty) string. In the future
// the interface may be broadened to make it behave more like a real
// type with methods to access commonly required information.
type SessionID string
