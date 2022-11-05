package main

import (
	"log"

	"github.com/karlpokus/mqtt-client/lib/stream"
)

func main() {
	exit := make(chan bool)
	req, res := stream.NewClient()
	go func() {
		for r := range res {
			if r.Notice() {
				if r.Fatal() {
					log.Printf("fatal %q", r.Err())
					exit <- true
					return
				}
				log.Printf("notice %q", r.Message)
				continue
			}
			log.Printf("message %q on topic %q", r.Message, r.Topic)
		}
	}()
	req <- stream.Subscribe("test")
	//req <- stream.Publish("test", []byte("hello world"))
	<-exit
}
