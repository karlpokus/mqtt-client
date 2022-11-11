package stream

import "github.com/karlpokus/mqtt-client/lib/packet"

// fakeStream holds a list of acks and
// a counter of reads
type fakeStream struct {
	next  []fakeAck
	reads int
}

type fakeAck struct {
	name string
	id   uint16
}

// push pushes the ack to the list
func (fake *fakeStream) push(ack fakeAck) {
	fake.next = append(fake.next, ack)
}

// pop removes-, and returns the ack from the list
func (fake *fakeStream) pop() fakeAck {
	if len(fake.next) == 0 {
		return fakeAck{}
	}
	out := fake.next[0]
	fake.next = fake.next[1:]
	return out
}

// Read pops-, and copies acks into p
func (fake *fakeStream) Read(p []byte) (int, error) {
	ack := fake.pop()
	switch ack.name {
	case packet.CONNACK:
		return copy(p, packet.Connack()), nil
	case packet.PINGRESP:
		return copy(p, packet.PingResp()), nil
	case packet.SUBACK:
		return copy(p, packet.Suback(ack.id)), nil
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

// Write appends acks to the list based on p
func (fake *fakeStream) Write(p []byte) (int, error) {
	switch packet.Packet[p[0]] {
	case packet.CONNECT:
		fake.push(fakeAck{name: packet.CONNACK})
		return len(p), nil
	case packet.PINGREQ:
		fake.push(fakeAck{name: packet.PINGRESP})
		return len(p), nil
	case packet.SUBSCRIBE:
		fake.push(fakeAck{
			name: packet.SUBACK,
			id:   packet.ParseSubscribe(p),
		})
		return len(p), nil
	default:
		return 0, ErrConnClosed
	}
}
