package packet

import (
	"fmt"
	"io"
	"log"

	"github.com/karlpokus/mqtt-client/lib/stream"
)

var ControlPacket = map[uint8]string{
	0xd0: "PINGRESP",
}

var connackReturnCodeDesc = []string{
	"Connection accepted",
	"The Server does not support the level of the MQTT protocol requested by the Client",
	"The Client identifier is correct UTF-8 but not allowed by the Server",
	"The Network Connection has been made but the MQTT service is unavailable",
	"The data in the user name or password is malformed",
	"The Client is not authorized to connect",
}

// Connect sends CONNECT and expects CONNACK in return
func Connect(rwc chan func(io.ReadWriter) error) {
	rwc <- func(rw io.ReadWriter) error {
		_, err := rw.Write(connect("bixa")) // bixa the cat
		if err != nil {
			return err
		}
		var b [4]byte
		_, err = rw.Read(b[:])
		if err != nil {
			return err
		}
		return connackVerify(b)
	}
}

// connackVerify verifies the control code and return code of CONNACK
func connackVerify(b [4]byte) error {
	if b[0] != 0x20 {
		return fmt.Errorf("Error: server response is not CONNACK: %x", b)
	}
	// TODO: verify session present flag
	if b[3] != 0 {
		return fmt.Errorf("Error: connack return code: %s", connackReturnCodeDesc[b[3]])
	}
	return nil
}

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

func Disconnect(rwc chan func(io.ReadWriter) error) chan bool {
	release := make(chan bool)
	rwc <- func(rw io.ReadWriter) error {
		defer func() {
			log.Println("Disconnect end") // debug
			release <- true
		}()
		log.Println("Disconnect start") // debug
		_, err := rw.Write([]byte{0xe0, 0})
		return err
	}
	return release
}

// Ping sends PINGREQ and expects PINGRESP in return
func Ping(rwc chan func(io.ReadWriter) error) {
	rwc <- func(rw io.ReadWriter) error {
		defer log.Println("Ping end") // debug
		log.Println("Ping start")     // debug
		ping := []byte{0xc0, 0}
		_, err := rw.Write(ping)
		if err != nil {
			return err
		}
		err = stream.SetReadDeadline(rw, 5)
		if err != nil {
			return err
		}
		var b [2]byte
		_, err = rw.Read(b[:])
		if err != nil {
			if stream.Timeout(err) {
				// TODO: yield and retry
				return fmt.Errorf("Error: read timeout waiting for PINGRESP")
			}
			if stream.Closed(err) {
				return fmt.Errorf("connection closed by server")
			}
			return err
		}
		v, ok := ControlPacket[b[0]]
		if ok && v == "PINGRESP" {
			return nil
		}
		return fmt.Errorf("Error: server response in not PINGRESP: %x", b)
	}
}
