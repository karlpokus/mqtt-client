package stream

import (
	"io"
	"log"
	"net"
	"time"
)

var controlPacket = map[uint8]string{
	0x10: "CONNECT",
	0x20: "CONNACK",
	0xc0: "PINGREQ",
	0xd0: "PINGRESP",
	0xe0: "DISCONNECT",
}

// TODO: Stream -> stream - it does not need to be exposed
type Stream struct {
	rw io.ReadWriter
}

// Listen exposes a Stream as a ReadWriter to funcs on rwc
func Listen(rwc chan func(io.ReadWriter) error, fatal chan error) {
	stm, err := new()
	if err != nil {
		fatal <- err
		return
	}
	for fn := range rwc {
		err := fn(stm)
		if err != nil {
			fatal <- err
			return
		}
	}
}

// new returns a Stream
func new() (*Stream, error) {
	conn, err := net.Dial("tcp", "localhost:1883")
	if err != nil {
		return nil, err
	}
	return &Stream{
		rw: conn,
	}, nil
}

// Read performs the duties of an io.Reader,
// sets a read deadline if it finds an embedded net.Conn and
// logs the op code of the read packet
func (stm *Stream) Read(p []byte) (int, error) {
	ttl := 5
	if conn, ok := stm.rw.(net.Conn); ok {
		err := conn.SetReadDeadline(time.Now().Add(time.Duration(ttl) * time.Second))
		if err != nil {
			return 0, err
		}
	}
	n, err := stm.rw.Read(p)
	op := p[0]
	if v, ok := controlPacket[op]; ok {
		log.Printf("%s read", v)
	} else {
		log.Printf("unknown op %x read", op)
	}
	return n, err
}

// Write performs the duties of an io.Writer and
// logs the op code of the written packet
func (stm *Stream) Write(p []byte) (int, error) {
	op := p[0]
	if v, ok := controlPacket[op]; ok {
		log.Printf("%s written", v)
	} else {
		log.Printf("unknown op %x written", op)
	}
	return stm.rw.Write(p)
}

// Timeout returns true if err is a timeout
func Timeout(err error) bool {
	if terr, ok := err.(net.Error); ok && terr.Timeout() {
		return true
	}
	return false
}

// Closed returns true if err indicates that the stream is closed
func Closed(err error) bool {
	return err == io.EOF
}
