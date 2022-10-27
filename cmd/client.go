package main

import (
	"log"
	"io"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/karlpokus/mqtt-client/lib/packet"
)

func parse(conn net.Conn, fatal chan error) {
	var b [1024]byte
	for {
		n, err := conn.Read(b[:]) // blocking read
		if err != nil {
			fatal <- err
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

func write(conn net.Conn, fatal chan error, b []byte) {
	_, err := conn.Write(b)
	if err != nil {
		fatal <- err
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
	log.Println("connection ok")
	rwc := make(chan func(io.ReadWriter) error)
	fatal := make(chan error)
	exit := make(chan bool)
	go func() {
		for fn := range rwc {
			// note: some error handling options here that also affects flow control:
			// pass fatal to fn
			// let the wrapper func return the error
			// let fn return a release chan
			err := fn(conn)
			if err != nil {
				fatal <- err
				return // fatal
			}
		}
	}()
	go func() {
		select {
		case err := <-fatal:
			log.Printf("%s", err)
		case <-interrupt():
			packet.Disconnect(rwc)
			// note: consider closing the connection.
			// or perhaps check if server closed it
		}
		exit <-true
	}()
	packet.Connect(rwc)
	go parse(conn, fatal)
	go func() {
		for {
			time.Sleep(30 * time.Second) // half the set keep-alive
			packet.Ping(rwc)
			// TODO: verify pingresp
		}
	}()
	<-exit
	log.Println("client exiting")
}
