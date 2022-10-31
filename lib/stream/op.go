package stream

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/karlpokus/mqtt-client/lib/packet"
)

type Op func(io.ReadWriter) error

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
		return connackVerify(b)
	}
}

// connackVerify verifies the control code and return code of CONNACK
func connackVerify(b [4]byte) error {
	if b[0] != 0x20 {
		return fmt.Errorf("Error: server response is not CONNACK: %x", b)
	}
	// TODO: verify session present flag
	if b[3] != 0 {
		return fmt.Errorf("Error: connack return code: %s", packet.ConnackReturnCodeDesc[b[3]])
	}
	return nil
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
			if timeout(err) {
				// TODO: yield and retry
				return fmt.Errorf("Error: read timeout waiting for PINGRESP")
			}
			if closed(err) {
				return fmt.Errorf("connection closed by server")
			}
			return err
		}
		v, ok := packet.ControlPacket[b[0]]
		if ok && v == "PINGRESP" {
			return nil
		}
		return fmt.Errorf("Error: server response in not PINGRESP: %x", b)
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
			if timeout(err) {
				log.Println("parse read timeout") // debug
				return nil
			}
			if closed(err) {
				return fmt.Errorf("connection closed by server")
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

// timeout returns true if err is a timeout
func timeout(err error) bool {
	if terr, ok := err.(net.Error); ok && terr.Timeout() {
		return true
	}
	return false
}

// closed returns true if err indicates that the stream is closed
func closed(err error) bool {
	return err == io.EOF
}
