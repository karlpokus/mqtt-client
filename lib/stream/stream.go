package stream

import (
	"io"
	"net"
	"time"
)

type Stream struct {
	rw io.ReadWriter
}

// Listen exposes a Stream to funcs on rwc
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

func (stm *Stream) Read(p []byte) (int, error) {
	//log.Printf("read %x") p[0]
	return stm.rw.Read(p)
}

func (stm *Stream) Write(p []byte) (int, error) {
	return stm.rw.Write(p)
}

// SetReadDeadline sets a read deadline if rw has an embedded net.Conn
func SetReadDeadline(rw io.ReadWriter, i int) error {
	if stm, ok := rw.(*Stream); ok {
		if conn, ok := stm.rw.(net.Conn); ok {
			err := conn.SetReadDeadline(time.Now().Add(time.Duration(i) * time.Second))
			if err != nil {
				return err
			}
		}
	}
	return nil
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
