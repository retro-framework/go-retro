package events

type SetAvatar struct {
	ContentType string
	ImgData     []byte
}

func init() {
	Register(&SetAvatar{})
}
