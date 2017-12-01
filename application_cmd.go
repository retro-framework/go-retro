package main

// CmdDesc is an abstract description of a command, it has a method
// name and a few other bits and pieces. A resolver will turn a command
// desc into a command we can run. The resolver may lookup the "name"
// of a command to a foreign object using something like GRPC allowing
// the dropping-in of microservices "behind" the main app loop.
type CmdDesc interface {
	Name() string
	Path() string
	Args() ApplicationCmdArgs
}

type ApplicationCmdArgs interface{}
