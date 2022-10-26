package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/karlpokus/mqtt-client/lib/packet"
)

func parse(conn net.Conn, errc chan error) {
	var b [1024]byte
	for {
		n, err := conn.Read(b[:]) // blocking read
		if err != nil {
			errc <- err
			return
		}
		v, ok := packet.ControlPacket[b[0]]
		if ok {
			log.Printf("%s recieved", v)
		} else {
			log.Printf("%x", b[:n]) // dump hex
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

func interrupt() <-chan os.Signal {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	return sigc
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
	err = packet.Connect(conn)
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
			write(conn, errc, packet.PingReq())
			// TODO: verify pingresp
		}
	}()
	select {
	case err := <-errc:
		log.Printf("%s", err)
	case <-interrupt():
		log.Println("DISCONNECT send")
		write(conn, errc, packet.Disconnect()) // will block on errc
		// TODO: close chan? Or check if server closed it, for fun
	}
	log.Println("client exiting")
}
