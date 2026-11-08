# Hints

## play

Everything in `handle` after `MsgWelcomeName` and the join belongs in `play`:
the deferred logout and the command loop. At the top of the deferred function:

```go
if s.handingOver.Load() {
	return // the new server has the player now
}
```

## PrepareCopyover

```go
s.lmu.Lock()
lf, err := fileOf(s.listeners[0])
s.lmu.Unlock()
files := []*os.File{lf}
s.call(func() {
	state.Ticks = s.ticks
	for _, name := range s.playerNames() {
		p := s.players[name]
		f, ferr := fileOf(p.conn)
		// ...
		files = append(files, f)
		cp := CopyoverPlayer{Name: p.Name, /* ... */, File: len(files) - 1}
	}
})
```

## HandOver

Start from `Shutdown`: the same closing of listeners, waiting for handlers and
stopping the loop – but `s.handingOver.Store(true)` first, and in the game loop
`p.conn.Close()` for every connection instead of the goodbye and the deadline.

## adopt, the goroutine

```go
s.lmu.Lock()
s.handlers.Add(1)
s.lmu.Unlock()
go func() {
	defer s.handlers.Done()
	defer conn.Close()
	defer p.close()
	defer s.call(func() { delete(s.conns, p) })
	s.play(p)
}()
```

## main.go

The changes to `main.go`, roughly:

```go
// After a copyover the listener is inherited, not opened.
statePath := os.Getenv("MUD_COPYOVER")
os.Unsetenv("MUD_COPYOVER")
var l net.Listener
var err error
if statePath == "" {
	if l, err = net.Listen("tcp", *addr); err != nil { /* as before */ }
}
// ...
srv := mud.NewServerWithOptions(mud.Options{ /* as before */ OnCopyover: startNewServer})
if statePath != "" {
	if l, err = adoptCopyover(srv, statePath); err != nil {
		log.Fatalf("copyover: %v", err)
	}
}
// ...
// the Ctrl+C goroutine: select { case <-ctx.Done(): case <-srv.Done(): return }
// and after Serve (the server stops either way; ctx says which way):
<-srv.Done()
if ctx.Err() != nil { // Ctrl+C
	<-shutdownDone
	log.Print("bye")
} else {
	log.Print("handed over to the new server – bye")
}
```
