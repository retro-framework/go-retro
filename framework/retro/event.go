package retro

// Event interface may be any type which may carry any baggage it likes.
// It must serialize and deserialize cleanly for storage reasons.
type Event interface{}
