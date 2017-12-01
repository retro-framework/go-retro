package events

type SetDisplayName struct {
	Name string `json:"name"`
}

func init() {
	Register(&SetDisplayName{})
}
