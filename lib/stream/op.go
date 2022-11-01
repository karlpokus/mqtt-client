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

// Parse reads from rw until timeout
func Parse(ops chan Op) chan bool {
	release := make(chan bool)
	ops <- func(rw io.ReadWriter) error {
		defer func() {
			release <- true
		}()
		log.Println("parse start") // debug
		var b [64]byte
		n, err := rw.Read(b[:]) // blocking
		if err != nil {
			if errors.Is(err, ErrReadTimeout) {
				log.Println("parse read timeout") // debug
				return nil
			}
			return err
		}
		v, ok := packet.ControlPacket[b[0]]
		if ok {
			log.Printf("%s recieved", v)
		} else {
			log.Printf("%x", b[:n]) // dump hex
		}
		return nil
	}
	return release
}
