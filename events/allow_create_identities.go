package events

type AllowCreateIdentities struct{}

func init() {
	Register(&AllowCreateIdentities{})
}
