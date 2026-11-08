# Ledtrådar

## play

Allt i `handle` efter `MsgWelcomeName` och inträdet hör till `play`: den
uppskjutna utloggningen och kommandoloopen. Överst i den uppskjutna funktionen:

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

Utgå från `Shutdown`: samma stängning av lyssnare, väntan på hanterarna och
stopp av loopen – men `s.handingOver.Store(true)` först, och i spelloopen
`p.conn.Close()` för varje anslutning i stället för adjö och deadline.

## adopt, goroutinen

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

Ändringarna i `main.go`, ungefär:

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
