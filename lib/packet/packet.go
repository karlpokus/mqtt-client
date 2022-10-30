package packet

var ControlPacket = map[uint8]string{
	0x10: "CONNECT",
	0x20: "CONNACK",
	0xc0: "PINGREQ",
	0xd0: "PINGRESP",
	0xe0: "DISCONNECT",
}

var ConnackReturnCodeDesc = []string{
	"Connection accepted",
	"The Server does not support the level of the MQTT protocol requested by the Client",
	"The Client identifier is correct UTF-8 but not allowed by the Server",
	"The Network Connection has been made but the MQTT service is unavailable",
	"The data in the user name or password is malformed",
	"The Client is not authorized to connect",
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
