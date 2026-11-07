// Entry point of your MUD server. Run it with: ./mudschool run server
// (or `go run .` in this directory). Connect with: ./mudschool connect
package main

import (
	"flag"
	"log"
	"net"

	"mudserver/mud"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:4000", "address to listen on (use :4000 to let other computers connect)")
	worldPath := flag.String("world", "", "world file (JSON); empty = the built-in world")
	flag.Parse()

	world := mud.DefaultWorld()
	if *worldPath != "" {
		var err error
		if world, err = mud.LoadWorldFile(*worldPath); err != nil {
			log.Fatalf("cannot load the world:\n%v", err)
		}
	}
	log.Printf("the world has %d rooms", len(world.Rooms))

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("cannot listen on %s: %v", *addr, err)
	}
	log.Printf("MUD server listening on %s", l.Addr())

	srv := mud.NewServerWithWorld(world)
	log.Fatal(srv.Serve(l))
}
