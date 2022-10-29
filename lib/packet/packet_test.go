package packet

import (
	"bytes"
	"io"
	"testing"
)

// Connect will first write (append) a connect packet to buf
// and then read the connack packet
// After the connack packet has been read we can test the connect packet
func TestConnectOk(t *testing.T) {
	connack := []byte{0x20, 0, 0, 0}
	buf := bytes.NewBuffer(connack)
	rwc := make(chan func(io.ReadWriter) error)
	go Connect(rwc)
	fn := <-rwc
	err := fn(buf)
	if err != nil {
		t.Fatalf("connack error: %s", err)
	}
	c, err := buf.ReadByte()
	if err != nil {
		t.Fatalf("readbyte error: %s", err)
	}
	if c != 0x10 {
		t.Fatalf("connect error: %x", c)
	}
}

// TODO: TestConnectFail

func TestPingOk(t *testing.T) {
	pingreq := []byte{0xc0, 0}
	pingresp := []byte{0xd0, 0}
	buf := bytes.NewBuffer(pingresp)
	rwc := make(chan func(io.ReadWriter) error)
	go Ping(rwc)
	fn := <-rwc
	err := fn(buf)
	if err != nil {
		t.Fatalf("pingresp error: %s", err)
	}
	b := buf.Bytes()
	if !bytes.Equal(b, pingreq) {
		t.Fatalf("pingreq error: %x", b)
	}
}

// TODO: TestPingFail
