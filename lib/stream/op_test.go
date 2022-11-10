package stream

import (
	"bytes"
	"errors"
	"testing"

	"github.com/karlpokus/mqtt-client/lib/packet"
)

// Connect will first write (append) a connect packet to buf
// and then read the connack packet
// After the connack packet has been read we can test the connect packet
func TestConnectOk(t *testing.T) {
	buf := bytes.NewBuffer(packet.Connack())
	ops := make(chan op)
	go connect(ops)
	op := <-ops
	err := op(buf)
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

func TestConnectFail(t *testing.T) {
	notConnack := []byte{0xd0, 0, 0, 0}
	buf := bytes.NewBuffer(notConnack)
	ops := make(chan op)
	go connect(ops)
	op := <-ops
	err := op(buf)
	if !errors.Is(err, ErrBadPacket) {
		t.Fatalf("unexpected error: %s", err)
	}
}
