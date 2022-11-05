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

// listen exposes a stream as a ReadWriter to funcs on the ops channel
func listen(ctx context.Context, ops chan op, fatal chan error) {
	stm, err := new()
	if err != nil {
		fatal <- err
		return
	}
	defer log.Println("DEBUG: ops listener closed")
	for {
		select {
		case op := <-ops:
			err := op(stm)
			if err != nil {
				fatal <- err
			}
		case <-ctx.Done():
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
// sets a read deadline and logs the op code of the read packet
func (stm *stream) Read(p []byte) (int, error) {
	defer func() {
		if p[0] != 0 {
			// something was read
			if v, ok := packet.ControlPacket[p[0]]; ok {
				log.Printf("%s", v)
			} else {
				log.Printf("unknown op %x read", p[0])
			}
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
	pass := make(chan bool)
	t := time.NewTimer(time.Duration(ttl) * time.Second)
	// note: this will leak on timeout
	go func() {
		n, err = stm.rw.Read(p)
		pass <- true
	}()
	select {
	case <-t.C:
		err = ErrReadTimeout
	case <-pass:
		t.Stop()
	}
	return n, err
}

// Write performs the duties of an io.Writer and
// logs the op code of the written packet
func (stm *stream) Write(p []byte) (int, error) {
	op := p[0]
	if v, ok := packet.ControlPacket[op]; ok {
		log.Printf("%s", v)
	} else {
		log.Printf("unknown op %x written", op)
	}
	return stm.rw.Write(p)
}
