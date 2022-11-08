package stream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/karlpokus/mqtt-client/lib/packet"
)

type op func(io.ReadWriter) error

var (
	ErrBadPacket     = errors.New("unexpected packet")
	ErrBadReturnCode = errors.New("bad return code")
)

// connect writes a CONNECT packet and expects to read
// a CONNACK packet in return
func connect(ops chan op) {
	ops <- func(rw io.ReadWriter) error {
		_, err := rw.Write(packet.Connect("bixa")) // bixa the cat
		if err != nil {
			return err
		}
		b := make([]byte, 4)
		_, err = rw.Read(b)
		if err != nil {
			// TODO: yield and retry on timeout
			return err
		}
		if !packet.Is(b, packet.CONNACK) {
			return fmt.Errorf("%x %w", b, ErrBadPacket)
		}
		// TODO: verify session present flag
		// TODO: verify n > 3 first
		if b[3] != 0 {
			return fmt.Errorf("%s %w", packet.ConnackReturnCodeDesc[b[3]], ErrBadReturnCode)
		}
		return nil
	}
}

// ping writes a PINGREQ packet
func ping(ctx context.Context, ops chan op, acks *packet.Acks) {
	defer log.Println("  ping cancelled")
	fn := func(rw io.ReadWriter) error {
		_, err := rw.Write(packet.PingReq())
		if err != nil {
			return err
		}
		// We must not read expecting PINGRESP here
		// since we might get another *ACK
		<-acks.Push(&packet.Ack{
			TTL:    10,
			Packet: packet.PingResp(),
		})
		return nil
	}
	for {
		select {
		case ops <- fn:
		case <-ctx.Done():
			return
		}
		time.Sleep(10 * time.Second) // 1/6 of the keep-alive ttl
	}
}

func disconnect(ops chan op) chan bool {
	release := make(chan bool)
	ops <- func(rw io.ReadWriter) error {
		defer func() {
			release <- true
		}()
		_, err := rw.Write(packet.Disconnect())
		return err
	}
	return release
}

// read reads from rw and writes to res
func read(ctx context.Context, ops chan op, acks *packet.Acks, res chan *Response) {
	defer log.Println("  read cancelled")
	fn := func(rw io.ReadWriter) error {
		defer log.Println("  read end")
		log.Println("  read start")
		b := make([]byte, 64)
		n, err := rw.Read(b)
		if err != nil {
			if errors.Is(err, ErrReadTimeout) {
				return nil
			}
			return err
		}
		if packet.Is(b, packet.PUBLISH) {
			t, m := packet.ParsePublish(b[:n])
			res <- &Response{
				Topic:   t,
				Message: m,
			}
			return nil
		}
		ack := acks.Pop(b[:n])
		if ack != nil {
			log.Printf("  %x popped", b[:n])
			if packet.Is(b, packet.SUBACK) {
				res <- notice("subscription acked")
			}
		}
		return nil
	}
	for {
		select {
		case ops <- fn:
		case <-ctx.Done():
			return
		}
	}
}

func subscribe(ctx context.Context, ops chan op, acks *packet.Acks, topic string) {
	fn := func(rw io.ReadWriter) error {
		_, err := rw.Write(packet.Subscribe(topic))
		if err != nil {
			return err
		}
		// We must not read expecting SUBACK here
		// since we might get PUBLISH first
		<-acks.Push(&packet.Ack{
			TTL:    30,
			Packet: packet.Suback(),
		})
		return nil
	}
	select {
	case ops <- fn:
	case <-ctx.Done():
		log.Println("  subscribe cancelled")
		return
	}
}
