package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

func read(conn net.Conn, errc chan error) {
	for {
		var b [128]byte
		n, err := conn.Read(b[:]) // arr to slice
		if err != nil {
			errc <- err
			return
		}
		log.Printf("%d bytes read", n)
		for i := 0; i < n; i++ {
			log.Printf("%x", b[i]) // dump hex for now
		}
	}
}

func write(conn net.Conn, errc chan error, b []byte) {
	n, err := conn.Write(b)
	if err != nil {
		errc <- err
		return
	}
	log.Printf("%d bytes written", n)
}

func main() {
	log.SetFlags(0)
  log.Println("client started")
	conn, err := net.Dial("tcp", "localhost:1883")
	if err != nil {
		log.Fatal(err)
	}
	errc := make(chan error)
	go read(conn, errc)
	go write(conn, errc, connect("bixa")) // bixa the cat
	// premature but let's start sending pings
  // hoping connect is fast
	go func() {
		for {
			time.Sleep(30 * time.Second) // half the set keep-alive
			write(conn, errc, pingreq())
		}
	}()
	select {
	case err := <-errc:
		log.Printf("%s", err)
	case <-interrupt():
		write(conn, errc, disconnect()) // might block on errc
	}
  log.Println("client exiting")
}

func interrupt() <-chan os.Signal {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	return sigc
}
