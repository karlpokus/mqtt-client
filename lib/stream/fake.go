package stream

import "github.com/karlpokus/mqtt-client/lib/packet"

// fakeStream holds a list of acks and
// a counter of reads
type fakeStream struct {
	next  []string
	reads int
}

// push pushes the ack to the list
func (fake *fakeStream) push(s string) {
	fake.next = append(fake.next, s)
}

// pop removes-, and returns the ack from the list
func (fake *fakeStream) pop() string {
	if len(fake.next) == 0 {
		return ""
	}
	out := fake.next[0]
	fake.next = fake.next[1:]
	return out
}

// Read pops acks
func (fake *fakeStream) Read(p []byte) (int, error) {
	switch fake.pop() {
	case packet.CONNACK:
		return copy(p, packet.Connack()), nil
	case packet.PINGRESP:
		return copy(p, packet.PingResp()), nil
	case packet.SUBACK:
		return copy(p, packet.Suback()), nil
	default:
		fake.reads++
		// allow a few empty reads
		// since we don't know the order of reads and writes
		if fake.reads > 3 {
			return 0, ErrConnClosed
		}
		return 0, nil
	}
}

// Write appends acks
func (fake *fakeStream) Write(p []byte) (int, error) {
	switch packet.Packet[p[0]] {
	case packet.CONNECT:
		fake.push(packet.CONNACK)
		return len(p), nil
	case packet.PINGREQ:
		fake.push(packet.PINGRESP)
		return len(p), nil
	case packet.SUBSCRIBE:
		fake.push(packet.SUBACK)
		return len(p), nil
	default:
		return 0, ErrConnClosed
	}
}
