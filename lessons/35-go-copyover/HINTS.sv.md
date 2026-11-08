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
