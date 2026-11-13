// Entry point of your MUD server. Run it with: ./mudschool run server
// (or `go run .` in this directory). Connect with: ./mudschool connect
package main

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mudserver/mud"
)

// The web client's files (lesson 39), baked into the program like world.json.
//
//go:embed web
var webFiles embed.FS

func main() {
	// Flags and environment variables (lesson 42).
	cfg, err := loadConfig(os.Args[1:], os.LookupEnv)
	if err != nil {
		log.Fatal(err)
	}
	if cfg.HealthCheck { // Docker's HEALTHCHECK runs this
		if err := healthCheck(cfg.Web); err != nil {
			log.Fatal(err)
		}
		return
	}
	// Everything – log.Printf too – goes through slog from here on.
	slog.SetDefault(newLogger(cfg.Log, os.Stderr))

	// Ctrl+C – or SIGTERM, which is how `docker stop` asks (lesson 42): shut
	// down cleanly instead of just dying. Caught from the start, so that a
	// signal while the server is starting isn't lost either.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	world := mud.DefaultWorld()
	if cfg.World != "" {
		var err error
		if world, err = mud.LoadWorldFile(cfg.World); err != nil {
			log.Fatalf("cannot load the world:\n%v", err)
		}
	}
	log.Printf("the world has %d rooms", len(world.Rooms))

	// After a copyover (lesson 35) the listener is inherited, not opened.
	statePath := os.Getenv("MUD_COPYOVER")
	os.Unsetenv("MUD_COPYOVER")
	var l net.Listener
	if statePath == "" {
		if l, err = net.Listen("tcp", cfg.Addr); err != nil {
			log.Fatalf("cannot listen on %s: %v", cfg.Addr, err)
		}
		log.Printf("MUD server listening on %s", l.Addr())
	}

	var store *mud.Store
	if cfg.DB != "" {
		if store, err = mud.OpenStore(cfg.DB); err != nil {
			log.Fatalf("cannot open the database: %v", err)
		}
		defer store.Close()
		log.Printf("players are saved in %s", cfg.DB)
	}

	opts := mud.Options{
		World: world, Tick: cfg.Tick, Store: store,
		WorldPath: cfg.World, ReloadEvery: cfg.Reload, Admins: cfg.Admins,
		ScriptDir: cfg.Scripts, PluginDir: cfg.Plugins,
		OnCopyover: startNewServer,
	}
	if cfg.Web != "" {
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
	if cfg.Web != "" {
		files, err := fs.Sub(webFiles, "web")
		if err != nil {
			log.Fatal(err)
		}
		gw := mud.NewGateway(files)
		gw.TrustProxy = cfg.TrustProxy // lesson 42
		mux := http.NewServeMux()
		mux.Handle("/healthz", srv.HealthHandler()) // lesson 42
		mux.Handle("/", gw)
		web = &http.Server{Addr: cfg.Web, Handler: mux, ReadHeaderTimeout: 10 * time.Second}
		wl, err := net.Listen("tcp", cfg.Web)
		if err != nil {
			log.Fatalf("cannot listen on %s: %v", cfg.Web, err)
		}
		log.Printf("web client on http://%s (copyover is off while it is on)", wl.Addr())
		go web.Serve(wl)
		go srv.Serve(gw.Listener())
	}

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
	if ctx.Err() != nil { // Ctrl+C or SIGTERM
		<-shutdownDone
		log.Print("bye")
	} else {
		log.Print("handed over to the new server – bye")
	}
}
