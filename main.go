package main

import (
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

func parse(conn net.Conn, errc chan error) {
	for {
		var b [128]byte
		n, err := conn.Read(b[:]) // arr to slice
		if err != nil {
			errc <- err
			return
		}
		if n == 0 { // needed?
			continue
		}
		v, ok := packetType[b[0]]
		if ok {
			log.Printf("%s recieved", v)
		} else {
			for i := 0; i < n; i++ {
				log.Printf("%x", b[i]) // dump hex
			}
		}
	}
}

func write(conn net.Conn, errc chan error, b []byte) {
	_, err := conn.Write(b)
	if err != nil {
		errc <- err
		return
	}
}

func main() {
	//log.SetFlags(0)
	log.Println("client started")
	conn, err := net.Dial("tcp", "localhost:1883")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("tcp connection ok")
	log.Println("CONNECT send")
	err = connect(conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("CONNACK recieved")
	errc := make(chan error)
	go parse(conn, errc)
	go func() {
		for {
			time.Sleep(30 * time.Second) // half the set keep-alive
			log.Println("PINGREQ send")
			write(conn, errc, pingreqPacket())
			// TODO: verify pingresp
		}
	}()
	select {
	case err := <-errc:
		log.Printf("%s", err)
	case <-interrupt():
    log.Println("DISCONNECT send")
    write(conn, errc, disconnectPacket()) // will block on errc
	}
	log.Println("client exiting")
}

// connect sends a connectPacket and expects a connack in return
func connect(rw io.ReadWriter) error {
	_, err := rw.Write(connectPacket("bixa")) // bixa the cat
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

func interrupt() <-chan os.Signal {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	return sigc
}
