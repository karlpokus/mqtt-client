package stream

import (
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"time"

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
	Err            error
}

var ErrInterrupt = errors.New("interrupt signal")

func (r *Response) Notice() bool {
	return r.Topic == "NOTICE"
}

func (r *Response) Fatal() bool {
	return r.Err != nil
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
		Err:   err,
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
		}
		// we have 3 ways of ending up here:
		// 1 rw error (cancel has already been run)
		// 2 expired acks (push to fatal)
		// 3 interrupt
		// but it's ok to call cancel multiple times
		cancel()
		if !errors.Is(err, ErrConnClosed) {
			<-disconnect(ops)
		}
		time.Sleep(1 * time.Second) // graceful shutdown
		close(ops)
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
