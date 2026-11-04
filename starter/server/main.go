// Entry point of your MUD server. Run it with: ./mudschool run server
// (or `go run .` in this directory). Connect with: telnet localhost 4000
package main

import (
	"flag"
	"log"
	"net"

	"mudserver/mud"
)

func main() {
	addr := flag.String("addr", ":4000", "address to listen on")
	flag.Parse()

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("cannot listen on %s: %v", *addr, err)
	}
	log.Printf("MUD server listening on %s", l.Addr())

	srv := mud.NewServer()
	log.Fatal(srv.Serve(l))
}
