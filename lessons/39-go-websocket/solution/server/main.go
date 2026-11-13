// Entry point of your MUD server. Run it with: ./mudschool run server
// (or `go run .` in this directory). Connect with: ./mudschool connect
package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"mudserver/mud"
)

// The web client's files (lesson 39), baked into the program like world.json.
//
//go:embed web
var webFiles embed.FS

func main() {
	addr := flag.String("addr", "127.0.0.1:4000", "address to listen on (use :4000 to let other computers connect)")
	worldPath := flag.String("world", "", "world file (JSON); empty = the built-in world")
	tick := flag.Duration("tick", mud.DefaultTick, "how often the world advances")
	dbPath := flag.String("db", "mud.db", "the SQLite database where players are saved; empty = save nothing")
	reload := flag.Duration("reload", 2*time.Second, "how often the -world file is checked for changes (0 = never)")
	admins := flag.String("admin", "", "comma-separated names of administrators, e.g. -admin Anna,Bertil")
	scripts := flag.String("scripts", "", "a folder with .lua scripts (lesson 32); empty = no scripts")
	plugins := flag.String("plugins", "", "a folder with .wasm plugins (lesson 33); empty = no plugins")
	webAddr := flag.String("web", "127.0.0.1:8080", "address of the web client (lesson 39); empty = no web client")
	flag.Parse()

	world := mud.DefaultWorld()
	if *worldPath != "" {
		var err error
		if world, err = mud.LoadWorldFile(*worldPath); err != nil {
			log.Fatalf("cannot load the world:\n%v", err)
		}
	}
	log.Printf("the world has %d rooms", len(world.Rooms))

	// After a copyover (lesson 35) the listener is inherited, not opened.
	statePath := os.Getenv("MUD_COPYOVER")
	os.Unsetenv("MUD_COPYOVER")
	var l net.Listener
	var err error
	if statePath == "" {
		if l, err = net.Listen("tcp", *addr); err != nil {
			log.Fatalf("cannot listen on %s: %v", *addr, err)
		}
		log.Printf("MUD server listening on %s", l.Addr())
	}

	var store *mud.Store
	if *dbPath != "" {
		if store, err = mud.OpenStore(*dbPath); err != nil {
			log.Fatalf("cannot open the database: %v", err)
		}
		defer store.Close()
		log.Printf("players are saved in %s", *dbPath)
	}

	opts := mud.Options{
		World: world, Tick: *tick, Store: store,
		WorldPath: *worldPath, ReloadEvery: *reload, Admins: strings.Split(*admins, ","),
		ScriptDir: *scripts, PluginDir: *plugins,
		OnCopyover: startNewServer,
	}
	if *webAddr != "" {
		// A copyover can't hand over the web server or WebSockets (lesson 39).
		opts.OnCopyover = nil
	}
	srv := mud.NewServerWithOptions(opts)
	if statePath != "" {
		if l, err = adoptCopyover(srv, statePath); err != nil {
			log.Fatalf("copyover: %v", err)
		}
		log.Printf("copyover done – still listening on %s", l.Addr())
	}

	// The web client (lesson 39): HTTP for the page, WebSockets for the game.
	var web *http.Server
	if *webAddr != "" {
		files, err := fs.Sub(webFiles, "web")
		if err != nil {
			log.Fatal(err)
		}
		gw := mud.NewGateway(files)
		web = &http.Server{Addr: *webAddr, Handler: gw, ReadHeaderTimeout: 10 * time.Second}
		wl, err := net.Listen("tcp", *webAddr)
		if err != nil {
			log.Fatalf("cannot listen on %s: %v", *webAddr, err)
		}
		log.Printf("web client on http://%s (copyover is off while it is on)", wl.Addr())
		go web.Serve(wl)
		go srv.Serve(gw.Listener())
	}

	// Ctrl+C: shut down cleanly instead of just dying.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		select {
		case <-ctx.Done():
		case <-srv.Done(): // handed over: nothing left to shut down
			return
		}
		log.Print("shutting down ...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown: %v", err)
		}
		if web != nil {
			web.Shutdown(shutdownCtx) // lesson 39
		}
	}()

	if err := srv.Serve(l); !errors.Is(err, mud.ErrServerClosed) {
		log.Fatal(err)
	}
	// Serve returns as soon as the listener is closed; wait for Shutdown to
	// finish, so that every player has been said goodbye to (and saved) – or
	// for a copyover to have handed everything over.
	<-srv.Done()
	if ctx.Err() != nil { // Ctrl+C
		<-shutdownDone
		log.Print("bye")
	} else {
		log.Print("handed over to the new server – bye")
	}
}
