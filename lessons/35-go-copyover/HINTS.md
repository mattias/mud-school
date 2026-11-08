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
