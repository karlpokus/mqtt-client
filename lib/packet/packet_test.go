package packet

import (
	"bytes"
	"testing"
  "io"
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
		t.Fatalf("connect error: %s", err)
	}
}

// TODO: TestConnectFail
