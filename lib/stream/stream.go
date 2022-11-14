package stream

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"time"

	"github.com/karlpokus/mqtt-client/lib/packet"
)

type stream struct {
	rw io.ReadWriter
}

var (
	ErrReadTimeout = errors.New("read timeout")
	ErrConnClosed  = errors.New("connection closed")
)

// listen exposes an io.ReadWriter to funcs on the ops channel
func listen(cancel context.CancelFunc, rw io.ReadWriter, ops chan op, fatal chan error) {
	defer log.Println("  listener closed")
	stm := &stream{rw: rw}
	for fn := range ops {
		err := fn(stm)
		if err != nil {
			// note: if we move DISCONNECT here
			// we can return and have a single cancel call
			// after reading from fatal
			cancel()
			fatal <- err
		}
	}
}

// Read performs the duties of an io.Reader,
// sets a read deadline and logs the op code of the read packet
func (stm *stream) Read(p []byte) (int, error) {
	defer func() {
		if p[0] != 0 {
			// something was read
			log.Printf("< %s", packet.Packet(p))
		}
	}()
	ttl := 5
	var n int
	var err error
	// net.Conn
	if conn, ok := stm.rw.(net.Conn); ok {
		err = conn.SetReadDeadline(time.Now().Add(time.Duration(ttl) * time.Second))
		if err != nil {
			return n, err
		}
		n, err = conn.Read(p)
		if err != nil {
			if terr, ok := err.(net.Error); ok && terr.Timeout() {
				return n, ErrReadTimeout
			}
			if err == io.EOF {
				return n, ErrConnClosed
			}
		}
		return n, err
	}
	// io.ReadWriter
	return stm.rw.Read(p)
}

// Write performs the duties of an io.Writer and
// logs the op code of the written packet
func (stm *stream) Write(p []byte) (int, error) {
	log.Printf("> %s", packet.Packet(p))
	return stm.rw.Write(p)
}
