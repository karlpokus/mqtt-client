package stream

type Request struct {
	topic   string
	payload []byte
	kind    string
}
