package packet

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// TODO: add friendly name for pop feedback?
type Ack struct {
	TTL    int
	Packet []byte
	cancel chan bool
}

type Acks struct {
	m    sync.Map
	errc chan error
}

var ErrAckExpired = errors.New("ack ttl expired")

func NewAcks(errc chan error) *Acks {
	return &Acks{errc: errc}
}

func (acks *Acks) Push(ack *Ack) <-chan bool {
	// NOTE: Stopping a time.Timer does not close
	// time.Timer.C (we're not even allowed to close
	// it since it's read-only).
	// So stopping a time.Timer while we're reading from it
	// in another goroutine will leave the channel open
	// and leak that goroutine.
	// So let's use an explicit cancellation signal instead
	release := make(chan bool)
	cancel := make(chan bool)
	go func() {
		defer log.Printf("DEBUG: %s goroutine exiting", hex(ack.Packet))
		t := time.NewTimer(time.Duration(ack.TTL) * time.Second)
		release <- true
		select {
		case <-t.C:
			acks.errc <- fmt.Errorf("%s %w", hex(ack.Packet), ErrAckExpired)
			return
		case <-cancel:
			stopped := t.Stop()
			log.Printf("DEBUG: %s timer stopped %t", hex(ack.Packet), stopped)
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
