package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/karlpokus/mqtt-client/lib/packet"
)

// parse reads from rw until timeout
func parse(rwc chan func(io.ReadWriter) error) chan bool {
	setReadDeadline := func(rw io.ReadWriter) error {
		if conn, ok := rw.(net.Conn); ok {
			err := conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			if err != nil {
				return err
			}
		}
		return nil
	}
	release := make(chan bool)
	rwc <- func(rw io.ReadWriter) error {
		defer func() {
			release <- true
		}()
		log.Println("parse start") // debug
		setReadDeadline(rw)
		var b [64]byte
		n, err := rw.Read(b[:]) // blocking read
		if err != nil {
			if terr, ok := err.(net.Error); ok && terr.Timeout() {
				log.Println("parse read timeout") // debug
				return nil
			}
			if err == io.EOF {
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
			<-packet.Disconnect(rwc)
		}
		exit <- true
	}()
	packet.Connect(rwc)
	go func() {
		for {
			packet.Ping(rwc)
			time.Sleep(10 * time.Second) // 1/6 of the keep-alive deadline
		}
	}()
	go func() {
		for {
			<-parse(rwc) // temporary dump func
		}
	}()
	<-exit
	log.Println("client exiting")
}
