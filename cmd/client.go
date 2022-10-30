package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/karlpokus/mqtt-client/lib/stream"
)

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
			<-stream.Disconnect(rwc)
		}
		log.Println("client exiting")
		os.Exit(1)
	}()
	go stream.Listen(rwc, fatal)
	stream.Connect(rwc)
	go func() {
		for {
			stream.Ping(rwc)
			time.Sleep(10 * time.Second) // 1/6 of the keep-alive deadline
		}
	}()
	for {
		<-stream.Parse(rwc) // temporary dump func
	}
}
