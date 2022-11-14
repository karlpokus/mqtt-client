package packet

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"time"
)

const (
	CONNECT    = "CONNECT"
	CONNACK    = "CONNACK"
	PINGREQ    = "PINGREQ"
	PINGRESP   = "PINGRESP"
	DISCONNECT = "DISCONNECT"
	SUBSCRIBE  = "SUBSCRIBE"
	SUBACK     = "SUBACK"
	PUBLISH    = "PUBLISH"
)

var controlCode = map[uint8]string{
	0x10: CONNECT,
	0x20: CONNACK,
	0xc0: PINGREQ,
	0xd0: PINGRESP,
	0xe0: DISCONNECT,
	0x82: SUBSCRIBE,
	0x90: SUBACK,
	0x30: PUBLISH,
}

var ConnackReturnCodeDesc = []string{
	"Connection accepted",
	"The Server does not support the level of the MQTT protocol requested by the Client",
	"The Client identifier is correct UTF-8 but not allowed by the Server",
	"The Network Connection has been made but the MQTT service is unavailable",
	"The data in the user name or password is malformed",
	"The Client is not authorized to connect",
}

// Packet returns the controlpacket of b formatted as a string
func Packet(b []byte) string {
	if len(b) == 0 {
		return "UNKNOWN"
	}
	if v, ok := controlCode[b[0]]; ok {
		return v
	}
	return "UNKNOWN"
}

// Is returns true if the key matches the value in Packet
func Is(b []byte, s string) bool {
	return Packet(b) == s
}

/*
CONNECT

fixed header: 2 bytes

	1 controlpacket + reserved
	1 remaining length

variable header: 10 bytes

	2 protocol name length
	4 protocol name string
	1 protocol version
	1 connect flags
	2 keep-alive

payload: 2 bytes + payload string

	2 payload length
	n payload string
*/
func Connect(clientId string) []byte {
	var buf bytes.Buffer
	protoName := "MQTT"
	protoVersion := 4 // 3.1.1
	connectFlags := 2 // clean session
	keepAlive := 60   // seconds
	// fixed header
	buf.WriteByte(0x10)
	buf.WriteByte(uint8(10 + 2 + len(clientId)))
	// variable header
	buf.Write(spread16(uint16(len(protoName))))
	buf.Write([]byte(protoName))
	buf.WriteByte(uint8(protoVersion))
	buf.WriteByte(uint8(connectFlags))
	buf.Write(spread16(uint16(keepAlive)))
	// payload
	buf.Write(spread16(uint16(len(clientId))))
	buf.Write([]byte(clientId))
	return buf.Bytes()
}

// Connack returns a CONNACK packet
func Connack() []byte {
	return []byte{0x20, 0, 0, 0}
}

// Disconnect returns a DISCONNECT packet
func Disconnect() []byte {
	return []byte{0xe0, 0}
}

// PingReq returns a PINGREQ packet
func PingReq() []byte {
	return []byte{0xc0, 0}
}

// PingResp returns a PINGRESP packet
func PingResp() []byte {
	return []byte{0xd0, 0}
}

/*
SUBSCRIBE

fixed header: 2 bytes

	1 controlpacket + reserved
	1 remaining length

variable header: 2 bytes

	2 packet id

payload: 2 + topic + QoS per topic

	2 payload length
	n payload string
	1 QoS
*/
func Subscribe(topic string, id uint16) []byte {
	var buf bytes.Buffer
	// fixed header
	buf.WriteByte(0x82)
	buf.WriteByte(uint8(2 + 2 + len(topic) + 1))
	// variable header
	buf.Write(spread16(id))
	// payload
	buf.Write(spread16(uint16(len(topic))))
	buf.Write([]byte(topic))
	buf.WriteByte(0)
	return buf.Bytes()
}

/*
SUBACK

fixed header: 2 bytes

	1 controlpacket
	1 remaining length

variable header: 2 bytes

	2 packet id

payload: 1 byte

	1 return code

Allowed return codes:

	0x00 - Success - Maximum QoS 0
	0x01 - Success - Maximum QoS 1
	0x02 - Success - Maximum QoS 2
	0x80 - Failure
*/
func Suback(id uint16) []byte {
	var buf bytes.Buffer
	// fixed header
	buf.WriteByte(0x90)
	buf.WriteByte(3)
	// var header
	buf.Write(spread16(id))
	// payload
	buf.WriteByte(0)
	return buf.Bytes()
}

/*
PUBLISH

fixed header: 2 bytes

	controlpacket + flags: dup, QoS, retain
	remaining length

variable header: 2 + topic string

	  topic length
		topic string
		(packed id only included in QoS > 0)

payload: n bytes

response:

	QoS 0: None
	QoS 1: PUBACK
	QoS 2: PUBREC
*/
func Publish(topic string, payload []byte) []byte {
	var buf bytes.Buffer
	buf.WriteByte(0x30)
	buf.WriteByte(uint8(2 + len(topic) + len(payload)))
	buf.Write(spread16(uint16(len(topic))))
	buf.Write([]byte(topic))
	buf.Write(payload)
	return buf.Bytes()
}

// ParsePublish returns topic and message from b
func ParsePublish(b []byte) (string, string) {
	// TODO: check DUP
	// TODO: check packet id if QoS > 0
	topicStart := 4
	topicLen := int(b[3])
	topicEnd := topicStart + topicLen
	topic := string(b[topicStart:topicEnd])
	message := string(b[topicEnd:len(b)])
	return topic, message
}

// ParseSubscribe returns the packet id from b
func ParseSubscribe(b []byte) uint16 {
	return unspread16(b[2:4])
}

// Id generates a random id > 0 < 0xffff
func Id() uint16 {
	rand.Seed(time.Now().UnixNano())
	return uint16(rand.Intn(0xffff))
}

// spread16 spreads an uint16 over 2 bytes
func spread16(i uint16) []byte {
	// []byte{byte(i >> 8), byte(i & 0xff)}
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, i)
	return b
}

// unspread16 merges 2 bytes into an uint16
func unspread16(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}
