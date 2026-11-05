// Entry point of your MUD server. Run it with: ./mudschool run server
// (or `go run .` in this directory). Connect with: telnet 127.0.0.1 4000
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"mudserver/mud"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:4000", "address to listen on (use :4000 to let other computers connect)")
	worldPath := flag.String("world", "", "world file (JSON); empty = the built-in world")
	tick := flag.Duration("tick", mud.DefaultTick, "how often the world advances")
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

	srv := mud.NewServerWithOptions(mud.Options{World: world, Tick: *tick})

	// Ctrl+C: shut down cleanly instead of just dying.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		log.Print("shutting down ...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()

	if err := srv.Serve(l); !errors.Is(err, mud.ErrServerClosed) {
		log.Fatal(err)
	}
	// Serve returns as soon as the listener is closed; wait for Shutdown to
	// finish, so that every player has been said goodbye to (and saved).
	<-shutdownDone
	log.Print("bye")
}
