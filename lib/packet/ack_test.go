package packet

import (
	"errors"
	"testing"
)

func TestPushPop(t *testing.T) {
	acks := NewAcks(nil)
	a := &Ack{TTL: 10}
	<-acks.Push(a)
	b := acks.Pop(a.Packet)
	if a != b {
		t.Fatalf("%v should match %v", a, b)
	}
}

func TestTimeout(t *testing.T) {
	errc := make(chan error)
	acks := NewAcks(errc)
	ack := &Ack{TTL: 1}
	<-acks.Push(ack)
	err := <-errc
	if !errors.Is(err, ErrAckExpired) {
		t.Fatalf("%s should match %s", err, ErrAckExpired)
	}
}
