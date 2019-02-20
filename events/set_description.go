package events

type SetDescription struct {
	Desc string `json:"desc"`
}

func init() {
	Register(&SetDescription{})
}
