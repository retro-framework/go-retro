package events

type PriorSessionsListed struct{}

func init() {
	Register(&PriorSessionsListed{})
}
