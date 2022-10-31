package stream

import (
	"bytes"
	"errors"
	"testing"

	"github.com/karlpokus/mqtt-client/lib/packet"
)

// note: we want tests to use streams so we can test operation timeouts
// (when this is implemented for embedded io.ReadWriters other than net.Conn)
// but we need a data source like bytes.Buffer to easily inspect
// packets written - so let's try this by passing a buffer to fake

// Connect will first write (append) a connect packet to buf
// and then read the connack packet
// After the connack packet has been read we can test the connect packet
func TestConnectOk(t *testing.T) {
	buf := bytes.NewBuffer(packet.Connack())
	stm := fake(buf)
	ops := make(chan Op)
	go Connect(ops)
	op := <-ops
	err := op(stm)
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
	stm := fake(buf)
	ops := make(chan Op)
	go Connect(ops)
	op := <-ops
	err := op(stm)
	if !errors.Is(err, errorNotConnack) {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestPingOk(t *testing.T) {
	buf := bytes.NewBuffer(packet.PingResp())
	stm := fake(buf)
	ops := make(chan Op)
	go Ping(ops)
	op := <-ops
	err := op(stm)
	if err != nil {
		t.Fatalf("pingresp error: %s", err)
	}
	b := buf.Bytes()
	if !bytes.Equal(b, packet.PingReq()) {
		t.Fatalf("pingreq error: %x", b)
	}
}

func TestPingFail(t *testing.T) {
	notPingresp := []byte{0xe0, 0}
	buf := bytes.NewBuffer(notPingresp)
	stm := fake(buf)
	ops := make(chan Op)
	go Ping(ops)
	op := <-ops
	err := op(stm)
	if !errors.Is(err, errorNotPingResp) {
		t.Fatalf("unexpected error: %s", err)
	}
}
