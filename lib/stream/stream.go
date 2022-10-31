package stream

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/karlpokus/mqtt-client/lib/packet"
)

type stream struct {
	rw io.ReadWriter
}

// Listen exposes a stream as a ReadWriter to Op funcs on the ops channel
func Listen(ops chan Op, fatal chan error) {
	stm, err := new()
	if err != nil {
		fatal <- err
		return
	}
	for op := range ops {
		err := op(stm)
		if err != nil {
			fatal <- err
			return
		}
	}
}

// new returns a stream
func new() (*stream, error) {
	conn, err := net.Dial("tcp", "localhost:1883")
	if err != nil {
		return nil, err
	}
	return &stream{
		rw: conn,
	}, nil
}

// Read performs the duties of an io.Reader,
// sets a read deadline if it finds an embedded net.Conn and
// logs the op code of the read packet
func (stm *stream) Read(p []byte) (int, error) {
	ttl := 5
	if conn, ok := stm.rw.(net.Conn); ok {
		err := conn.SetReadDeadline(time.Now().Add(time.Duration(ttl) * time.Second))
		if err != nil {
			return 0, err
		}
	}
	n, err := stm.rw.Read(p)
	op := p[0]
	if v, ok := packet.ControlPacket[op]; ok {
		log.Printf("%s read", v)
	} else {
		log.Printf("unknown op %x read", op)
	}
	return n, err
}

// Write performs the duties of an io.Writer and
// logs the op code of the written packet
func (stm *stream) Write(p []byte) (int, error) {
	op := p[0]
	if v, ok := packet.ControlPacket[op]; ok {
		log.Printf("%s written", v)
	} else {
		log.Printf("unknown op %x written", op)
	}
	return stm.rw.Write(p)
}
