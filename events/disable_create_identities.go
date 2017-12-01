package events

type DisableCreateIdentities struct{}

func init() {
	Register(&DisableCreateIdentities{})
}
