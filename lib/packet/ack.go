package packet

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

type Ack struct {
	TTL    int
	Packet []byte
	Name   string
	cancel chan bool
}

type Acks struct {
	m    sync.Map
	errc chan error
}

var ErrAckExpired = errors.New("ttl expired")

func NewAcks(errc chan error) *Acks {
	return &Acks{errc: errc}
}

func (acks *Acks) Push(ack *Ack) <-chan bool {
	if len(ack.Packet) > 0 {
		ack.Name = Packet[ack.Packet[0]]
	}
	release := make(chan bool)
	cancel := make(chan bool)
	go func() {
		var stopped bool
		defer func() {
			log.Printf("  %s timer exiting. Timer stopped %t", ack.Name, stopped)
		}()
		t := time.NewTimer(time.Duration(ack.TTL) * time.Second)
		release <- true
		select {
		case <-t.C:
			acks.errc <- fmt.Errorf("%s %w", ack.Name, ErrAckExpired)
			return
		case <-cancel:
			stopped = t.Stop()
		}
	}()
	ack.cancel = cancel
	acks.m.Store(hex(ack.Packet), ack)
	return release
}

// returning non-nil Ack means item found (and popped)
func (acks *Acks) Pop(b []byte) *Ack {
	if v, ok := acks.m.LoadAndDelete(hex(b)); ok {
		ack := v.(*Ack)
		ack.cancel <- true
		return ack
	}
	return nil
}

func hex(b []byte) string {
	return fmt.Sprintf("%x", b)
}
