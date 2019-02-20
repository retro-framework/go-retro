package events

type CreateListingImage struct {
	Data []byte `json:"data"`
}

func init() {
	Register(&CreateListingImage{})
}
