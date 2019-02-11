package events

type SetAvatar struct {
	ContentType string `json:"contentType"`
	ImgData     []byte `json:"imgData"`
}

func init() {
	Register(&SetAvatar{})
}
