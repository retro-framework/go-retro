package events

type StartSession struct{}

func init() {
	Register(&StartSession{})
}
