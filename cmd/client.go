package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/karlpokus/mqtt-client/lib/packet"
	"github.com/karlpokus/mqtt-client/lib/stream"
)

// parse reads from rw until timeout
func parse(rwc chan func(io.ReadWriter) error) chan bool {
	release := make(chan bool)
	rwc <- func(rw io.ReadWriter) error {
		defer func() {
			release <- true
		}()
		log.Println("parse start") // debug
		err := stream.SetReadDeadline(rw, 5)
		if err != nil {
			return err
		}
		var b [64]byte
		n, err := rw.Read(b[:]) // blocking read
		if err != nil {
			if stream.Timeout(err) {
				log.Println("parse read timeout") // debug
				return nil
			}
			if stream.Closed(err) {
				return fmt.Errorf("connection closed by server")
			}
			return err
		}
		v, ok := packet.ControlPacket[b[0]]
		if ok {
			log.Printf("%s recieved", v)
		} else {
			log.Printf("%x", b[:n]) // dump hex
		}
		return nil
	}
	return release
}

func interrupt() <-chan os.Signal {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	return sigc
}

func main() {
	//log.SetFlags(0)
	log.Println("client started")
	fatal := make(chan error)
	rwc := make(chan func(io.ReadWriter) error)
	go func() {
		select {
		case err := <-fatal:
			log.Printf("%s", err)
		case <-interrupt():
			<-packet.Disconnect(rwc)
		}
		log.Println("client exiting")
		os.Exit(1)
	}()
	go stream.Listen(rwc, fatal)
	packet.Connect(rwc)
	go func() {
		for {
			packet.Ping(rwc)
			time.Sleep(10 * time.Second) // 1/6 of the keep-alive deadline
		}
	}()
	for {
		<-parse(rwc) // temporary dump func
	}
}
