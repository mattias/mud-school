// Startpunkten för din MUD-server. Kör med: ./mudschool run server
// (eller `go run .` i den här katalogen). Anslut med: telnet localhost 4000
package main

import (
	"flag"
	"log"
	"net"

	"mudserver/mud"
)

func main() {
	addr := flag.String("addr", ":4000", "adress att lyssna på")
	flag.Parse()

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("kunde inte lyssna på %s: %v", *addr, err)
	}
	log.Printf("MUD-servern lyssnar på %s", l.Addr())

	srv := mud.NewServer()
	log.Fatal(srv.Serve(l))
}
