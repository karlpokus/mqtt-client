package stream

import (
	"bytes"
	"io"
	"testing"

	"github.com/karlpokus/mqtt-client/lib/packet"
)

// Connect will first write (append) a connect packet to buf
// and then read the connack packet
// After the connack packet has been read we can test the connect packet
func TestConnectOk(t *testing.T) {
	buf := bytes.NewBuffer(packet.Connack())
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
	buf := bytes.NewBuffer(packet.PingResp())
	rwc := make(chan func(io.ReadWriter) error)
	go Ping(rwc)
	fn := <-rwc
	err := fn(buf)
	if err != nil {
		t.Fatalf("pingresp error: %s", err)
	}
	b := buf.Bytes()
	if !bytes.Equal(b, packet.PingReq()) {
		t.Fatalf("pingreq error: %x", b)
	}
}

// TODO: TestPingFail
