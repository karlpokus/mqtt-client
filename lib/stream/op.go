package stream

import (
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/karlpokus/mqtt-client/lib/packet"
)

type Op func(io.ReadWriter) error

var (
	ErrBadPacket     = errors.New("unexpected packet")
	ErrBadReturnCode = errors.New("bad return code")
)

// Connect sends CONNECT and expects CONNACK in return
func Connect(ops chan Op) {
	ops <- func(rw io.ReadWriter) error {
		_, err := rw.Write(packet.Connect("bixa")) // bixa the cat
		if err != nil {
			return err
		}
		var b [4]byte
		_, err = rw.Read(b[:])
		if err != nil {
			return err
		}
		if !packet.Is(b[0], "CONNACK") {
			return fmt.Errorf("%x %w", b, ErrBadPacket)
		}
		// TODO: verify session present flag
		if b[3] != 0 {
			return fmt.Errorf("%s %w", packet.ConnackReturnCodeDesc[b[3]], ErrBadReturnCode)
		}
		return nil
	}
}

// Ping sends PINGREQ and expects PINGRESP in return
func Ping(ops chan Op) {
	ops <- func(rw io.ReadWriter) error {
		_, err := rw.Write(packet.PingReq())
		if err != nil {
			return err
		}
		var b [2]byte
		_, err = rw.Read(b[:])
		if err != nil {
			// TODO: yield and retry if timeout
			// if errors.Is(err, ErrReadTimeout) {}
			return err
		}
		if !packet.Is(b[0], "PINGRESP") {
			return fmt.Errorf("%x %w", b, ErrBadPacket)
		}
		return nil
	}
}

func Disconnect(ops chan Op) chan bool {
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

// Read reads from rw and writes to res
func Read(ops chan Op, res chan *Response) chan bool {
	release := make(chan bool)
	ops <- func(rw io.ReadWriter) error {
		defer func() {
			log.Println("parse end") // debug
			release <- true
		}()
		log.Println("parse start") // debug
		var b [64]byte
		n, err := rw.Read(b[:]) // blocking
		if err != nil {
			if errors.Is(err, ErrReadTimeout) {
				return nil
			}
			return err
		}
		if packet.Is(b[0], "PUBLISH") {
			t, m := packet.ParsePublish(b[:n])
			res <- &Response{
				topic:   t,
				message: m,
			}
		}
		return nil
	}
	return release
}

func Subscribe(ops chan Op, topic string) {
	ops <- func(rw io.ReadWriter) error {
		_, err := rw.Write(packet.Subscribe(topic))
		if err != nil {
			return err
		}
		var b [5]byte
		_, err = rw.Read(b[:])
		if err != nil {
			// TODO: yield and retry if timeout
			return err
		}
		if !packet.Is(b[0], "SUBACK") {
			return fmt.Errorf("%x %w", b, ErrBadPacket)
		}
		// TODO: check return code in b[5]
		return nil
	}
}
