package retro

// IDFn is a function that should generate IDs. This is primarily
// used in the Engine implementations to generate an ID for newly created
// things where no ID has been provided.
type IDFn func() (string, error)
