package stream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/karlpokus/mqtt-client/lib/packet"
)

var ErrInterrupt = errors.New("interrupt signal")

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
			switch r.kind {
			case "subscribe":
				subscribe(ctx, ops, acks, r.topic)
			case "publish":
				publish(ctx, ops, r.topic, r.payload)
			case "disconnect":
				fatal <- fmt.Errorf("client initiated disconnect")
			}
		}
	}()
	go ping(ctx, ops, acks)
	go read(ctx, ops, acks, res)
	return req, res
}

func interrupt() <-chan os.Signal {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	return sigc
}
