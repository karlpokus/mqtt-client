package packet

import "bytes"

var ControlPacket = map[uint8]string{
	0x10: "CONNECT",
	0x20: "CONNACK",
	0xc0: "PINGREQ",
	0xd0: "PINGRESP",
	0xe0: "DISCONNECT",
	0x82: "SUBSCRIBE",
	0x90: "SUBACK",
	0x30: "PUBLISH",
}

var ConnackReturnCodeDesc = []string{
	"Connection accepted",
	"The Server does not support the level of the MQTT protocol requested by the Client",
	"The Client identifier is correct UTF-8 but not allowed by the Server",
	"The Network Connection has been made but the MQTT service is unavailable",
	"The data in the user name or password is malformed",
	"The Client is not authorized to connect",
}

// Is returns true if key b matches value s in ControlPacket
func Is(b uint8, s string) bool {
	if v, ok := ControlPacket[b]; ok && v == s {
		return true
	}
	return false
}

// connect returns a CONNECT packet
func Connect(clientId string) []byte {
	fixedHeaderLen := 2
	varHeaderLen := 10
	payloadLen := 2 + len(clientId)
	totalLength := fixedHeaderLen + varHeaderLen + payloadLen
	// subtract clientId from totalLength so we may easily append it at the end
	b := make([]byte, totalLength-len(clientId))
	// fixed header: control packet type and reserved bits
	b[0] = 0x10
	// fixed header: remaining length (TODO: read 2.2.3 for proper implementation)
	b[1] = uint8(varHeaderLen + payloadLen)
	// variable header:
	b[2] = 0
	b[3] = 4
	b[4] = 0x4d // M
	b[5] = 0x51 // Q
	b[6] = 0x54 // T
	b[7] = 0x54 // T
	// variable header: proto version
	b[8] = 4
	// variable header: connect flags (set to clean session)
	b[9] = 2
	// variable header: keep-alive
	b[10] = 0
	b[11] = 60 // 0x3c is 60s
	// payload: length
	b[12] = 0
	b[13] = uint8(len(clientId))
	// payload: clientId
	b = append(b, []byte(clientId)...)
	return b
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

// fixed var   payload
// 82 09 00 01 00 04 74 65    73 74 00
// .  .  .  .  .  .  t  e     s  t  .
func Subscribe(topic string) []byte {
	var buf bytes.Buffer
	// fixed header - 2 bytes
	buf.WriteByte(0x82)
	buf.WriteByte(uint8(2 + 2 + len(topic) + 1))
	// var header - 2 bytes
	buf.WriteByte(0) // packet id MSB
	buf.WriteByte(1) // packet id LSB TODO: need to bump for future subscriptions?
	// payload
	buf.WriteByte(0)                 // len MSB
	buf.WriteByte(uint8(len(topic))) // len LSB
	buf.Write([]byte(topic))
	buf.WriteByte(0) // QoS per topic
	return buf.Bytes()
}

// fixed var   payload
// 90 03 00 01 00
// TODO: compare packet id
// note: we might get PUBLISH before SUBACK
func Suback() []byte {
	var buf bytes.Buffer
	// fixed header - 2 bytes
	buf.WriteByte(0x90)
	buf.WriteByte(3) // remaining len
	// var header - 2 bytes
	buf.WriteByte(0) // packet id MSB
	buf.WriteByte(1) // packet id LSB
	// payload - 1 byte
	// Allowed return codes:
	// 0x00 - Success - Maximum QoS 0
	// 0x01 - Success - Maximum QoS 1
	// 0x02 - Success - Maximum QoS 2
	// 0x80 - Failure
	buf.WriteByte(0)
	return buf.Bytes()
}

// pub: svammel
// fixed var               payload
// 30 0d 00 04 74 65 73 74 73 76 61 6d 6d 65 6c
//	           t  e  s  t  s  v  a  m  m  e  l
// TODO: check DUP
// TODO: check packet id
// ParsePublish returns topic and message from b
func ParsePublish(b []byte) (string, string) {
	// fixed header: 2 bytes
	// var header: 2 + len(topic) bytes
	// no packet id?
	topicStart := 4
	topicLen := int(b[3])
	topicEnd := topicStart + topicLen
	topic := string(b[topicStart:topicEnd])
	message := string(b[topicEnd:len(b)])
	return topic, message
}
