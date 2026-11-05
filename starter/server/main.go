// Entry point of your MUD server. Run it with: ./mudschool run server
// (or `go run .` in this directory). Connect with: telnet 127.0.0.1 4000
package main

import (
	"flag"
	"log"
	"net"

	"mudserver/mud"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:4000", "address to listen on (use :4000 to let other computers connect)")
	flag.Parse()

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("cannot listen on %s: %v", *addr, err)
	}
	log.Printf("MUD server listening on %s", l.Addr())

	srv := mud.NewServer()
	log.Fatal(srv.Serve(l))
}
