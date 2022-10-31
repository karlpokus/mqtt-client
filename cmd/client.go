package main

import (
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
	ops := make(chan stream.Op)
	go func() {
		select {
		case err := <-fatal:
			log.Printf("%s", err)
		case <-interrupt():
			<-stream.Disconnect(ops)
		}
		log.Println("client exiting")
		os.Exit(1)
	}()
	go stream.Listen(ops, fatal)
	stream.Connect(ops)
	go func() {
		for {
			stream.Ping(ops)
			time.Sleep(10 * time.Second) // 1/6 of the keep-alive deadline
		}
	}()
	for {
		<-stream.Parse(ops) // temporary dump func
	}
}
