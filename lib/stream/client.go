package stream

import (
	"context"
	"errors"
	"fmt"
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

func NewClient() (chan<- *Request, <-chan *Response) {
	//log.SetFlags(0)
	fatal := make(chan error)
	ops := make(chan op)
	req := make(chan *Request)
	res := make(chan *Response)
	acks := packet.NewAcks(fatal)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		var err error
		defer cancel()
		select {
		case err = <-fatal:
		case <-interrupt():
			err = ErrInterrupt
		}
		if !errors.Is(err, ErrConnClosed) {
			<-disconnect(ops)
		}
		res <- noticeFatal(err)
		// TODO: close connection?
	}()
	go listen(ctx, ops, fatal)
	connect(ops)
	go func() {
		for r := range req {
			if isSub(r) {
				res <- notice(fmt.Sprintf("subscribed to %s", r.topic))
				subscribe(ops, acks, r.topic)
				continue
			}
			/* pub echo
			res <- &Response{
				topic:   "test",
				message: string(r.payload),
			}*/
		}
	}()
	go func() {
		for {
			// TODO: retry on timeout
			// return ttl on channel
			ping(ops)
			time.Sleep(10 * time.Second) // 1/6 of the keep-alive deadline
		}
	}()
	go func() {
		for {
			<-read(ops, acks, res)
		}
	}()
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
