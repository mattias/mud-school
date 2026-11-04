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
	worldPath := flag.String("world", "", "världsfil (JSON); tom = den inbyggda världen")
	flag.Parse()

	world := mud.DefaultWorld()
	if *worldPath != "" {
		var err error
		if world, err = mud.LoadWorldFile(*worldPath); err != nil {
			log.Fatalf("kunde inte läsa världen:\n%v", err)
		}
	}
	log.Printf("världen har %d rum", len(world.Rooms))

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("kunde inte lyssna på %s: %v", *addr, err)
	}
	log.Printf("MUD-servern lyssnar på %s", l.Addr())

	srv := mud.NewServerWithWorld(world)
	log.Fatal(srv.Serve(l))
}
