package retro

// CommandDesc is used after parsing the raw byte stream from downstream
// clients. A CommandDesc is used internally as soon as the JSON parsing
// is done by the superficial request handling layer.
type CommandDesc interface {
	Name() string
	Path() string
}
