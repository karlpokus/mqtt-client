package main

func connect(clientId string) []byte {
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

func disconnect() []byte {
	b := make([]byte, 2)
	b[0] = 0xe0
	b[1] = 0
	return b
}

func pingreq() []byte {
	b := make([]byte, 2)
	b[0] = 0xc0
	b[1] = 0
	return b
}
