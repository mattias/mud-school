// Entry point of your MUD server. Run it with: ./mudschool run server
// (or `go run .` in this directory). Connect with: ./mudschool connect
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"mudserver/mud"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:4000", "address to listen on (use :4000 to let other computers connect)")
	worldPath := flag.String("world", "", "world file (JSON); empty = the built-in world")
	tick := flag.Duration("tick", mud.DefaultTick, "how often the world advances")
	dbPath := flag.String("db", "mud.db", "the SQLite database where players are saved; empty = save nothing")
	reload := flag.Duration("reload", 2*time.Second, "how often the -world file is checked for changes (0 = never)")
	admins := flag.String("admin", "", "comma-separated names of administrators, e.g. -admin Anna,Bertil")
	scripts := flag.String("scripts", "", "a folder with .lua scripts (lesson 32); empty = no scripts")
	plugins := flag.String("plugins", "", "a folder with .wasm plugins (lesson 33); empty = no plugins")
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

	srv := mud.NewServerWithOptions(mud.Options{
		World: world, Tick: *tick, Store: store,
		WorldPath: *worldPath, ReloadEvery: *reload, Admins: strings.Split(*admins, ","),
		ScriptDir: *scripts, PluginDir: *plugins,
		OnCopyover: startNewServer,
	})
	if statePath != "" {
		if l, err = adoptCopyover(srv, statePath); err != nil {
			log.Fatalf("copyover: %v", err)
		}
		log.Printf("copyover done – still listening on %s", l.Addr())
	}

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
	// finish, so that every player has been said goodbye to (and saved) – or
	// for a copyover to have handed everything over.
	select {
	case <-shutdownDone:
		log.Print("bye")
	case <-srv.Done():
		log.Print("handed over to the new server – bye")
	}
}

// startNewServer starts this program again (lesson 35) – the executable on
// disk, so a freshly built binary means new code – with the listener and the
// players' connections as extra files (fd 3, 4, ...) and the state in a file.
func startNewServer(state []byte, files []*os.File) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	f, err := os.CreateTemp("", "mud-copyover-*.json")
	if err != nil {
		return err
	}
	if _, err := f.Write(state); err != nil {
		f.Close()
		return err
	}
	f.Close()
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.ExtraFiles = files
	cmd.Env = append(os.Environ(), "MUD_COPYOVER="+f.Name())
	return cmd.Start()
}

// adoptCopyover reads the state and takes over the inherited files.
func adoptCopyover(srv *mud.Server, statePath string) (net.Listener, error) {
	data, err := os.ReadFile(statePath)
	os.Remove(statePath)
	if err != nil {
		return nil, err
	}
	var state mud.CopyoverState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	// ExtraFiles[k] becomes file descriptor 3+k in the new process.
	n := 1
	for _, p := range state.Players {
		n = max(n, p.File+1)
	}
	files := make([]*os.File, n)
	for k := range files {
		files[k] = os.NewFile(uintptr(3+k), fmt.Sprintf("copyover-%d", k))
	}
	return srv.Adopt(&state, files)
}
