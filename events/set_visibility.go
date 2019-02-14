package events

type SetVisibility struct {
	Radius string `json:"visibility"`
}

func init() {
	Register(&SetVisibility{})
}
