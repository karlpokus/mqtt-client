package stream

import (
	"context"
	"errors"
	"io"
	"os"
	"os/signal"

	"github.com/karlpokus/mqtt-client/lib/packet"
)

// Request is a subscription if payload is nil
// otherwise a publication
type Request struct {
	topic   string
	payload []byte
}

type Response struct {
	Topic, Message string
	err            error
}

var ErrInterrupt = errors.New("interrupt signal")

func (r *Response) Notice() bool {
	return r.Topic == "NOTICE"
}

func (r *Response) Fatal() bool {
	return r.err != nil
}

func (r *Response) Err() string {
	return r.err.Error()
}

// user-friendly wrapper
func Subscribe(topic string) *Request {
	return &Request{topic: topic}
}

// user-friendly wrapper
func Publish(topic string, payload []byte) *Request {
	return &Request{topic, payload}
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
		err:   err,
	}
}

func Client(rw io.ReadWriter) (chan<- *Request, <-chan *Response) {
	// note: wrap everything below in a parent goroutine to return asap
	fatal := make(chan error)
	ops := make(chan op)
	req := make(chan *Request)
	res := make(chan *Response)
	acks := packet.NewAcks(fatal)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		var err error
		select {
		case err = <-fatal:
		case <-interrupt():
			err = ErrInterrupt
			cancel()
		}
		// At this point cancel has been run
		if !errors.Is(err, ErrConnClosed) {
			<-disconnect(ops)
		}
		close(ops) // so this should be safe
		res <- noticeFatal(err)
	}()
	go listen(cancel, rw, ops, fatal)
	connect(ops)
	go func() {
		for r := range req {
			if isSub(r) {
				subscribe(ctx, ops, acks, r.topic)
				continue
			}
			/* pub echo
			res <- &Response{
				topic:   "test",
				message: string(r.payload),
			}*/
		}
	}()
	go ping(ctx, ops, acks)
	go read(ctx, ops, acks, res)
	return req, res
}

func isSub(r *Request) bool {
	return r.payload == nil
}

func interrupt() <-chan os.Signal {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	return sigc
}
