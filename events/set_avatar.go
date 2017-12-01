package events

type SetAvatar struct {
	ContentType string `json:"contentType"`
	ImgData     []byte `json:"imageData"`
}

func init() {
	Register(&SetAvatar{})
}
