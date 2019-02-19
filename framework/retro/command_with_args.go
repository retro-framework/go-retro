package retro

// CommandWithArgs is an optional interface upgrade on Command
// which exposes a new SetArgs command which can be used to pass
// user data (parsed out of {params: ...} in the CommandDesc)
type CommandWithArgs interface {
	Command
	SetArgs(CommandArgs) error
}
