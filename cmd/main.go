package main

import (
	"log"
	"net"

	"github.com/karlpokus/mqtt-client/lib/stream"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:1883")
	if err != nil {
		log.Fatal(err)
	}
	exit := make(chan bool)
	req, res := stream.Client(conn)
	go func() {
		for r := range res {
			if r.Notice() {
				if r.Fatal() {
					log.Printf("u %s", r.Err)
					exit <- true
					return
				}
				log.Printf("u %s", r.Message)
				continue
			}
			log.Printf("u %s on topic %s", r.Message, r.Topic)
		}
	}()
	req <- stream.Subscribe("test")
	req <- stream.Publish("test", []byte("hello w0rld"))
	//req <- stream.Disconnect()
	<-exit
}
