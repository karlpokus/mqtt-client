package stream

type Response struct {
	Topic, Message string
	Err            error
}

func (r *Response) Notice() bool {
	return r.Topic == "NOTICE"
}

func (r *Response) Fatal() bool {
	return r.Err != nil
}

// user-friendly wrapper
func Subscribe(topic string) *Request {
	return &Request{topic: topic, kind: "subscribe"}
}

// user-friendly wrapper
func Publish(topic string, payload []byte) *Request {
	return &Request{topic: topic, payload: payload, kind: "publish"}
}

// user-friendly wrapper
func Disconnect() *Request {
	return &Request{kind: "disconnect"}
}

func notice(m string) *Response {
	return &Response{
		Topic:   "NOTICE",
		Message: m,
	}
}

func noticeFatal(err error) *Response {
	return &Response{
		Topic: "NOTICE",
		Err:   err,
	}
}
