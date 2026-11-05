# Hints

## The loop

```go
ticker := time.NewTicker(s.tick)
defer ticker.Stop()
for {
    select {
    case f := <-s.events:
        f()
    case <-ticker.C:
        s.ticks++
        s.onTick()
    case <-s.stop:
        return
    }
}
```

## handle with call

```go
p := newPlayer(conn)
defer p.close()
if !s.call(func() { s.conns[p] = true }) {
    return
}
defer s.call(func() { delete(s.conns, p) })
...
for !p.quit {
    line, ok := p.readLine()
    if !ok {
        return
    }
    if !s.call(func() { s.dispatch(p, line) }) {
        return
    }
}
```

Defers run in reverse order: the player is removed from `s.players` and
`s.conns` *before* `p.close()` – so nobody can send to a closed outbox.

## Shutdown

```go
s.lmu.Lock()                 // the small mutex that protects s.listeners
s.closing.Store(true)        // atomic.Bool – Serve checks it after Accept fails
for _, l := range s.listeners { l.Close() }
s.lmu.Unlock()
s.call(func() {
    for p := range s.conns {
        p.Send(MsgShutdown)
        p.conn.SetReadDeadline(time.Now())
    }
})
// wait for s.handlers (a sync.WaitGroup, Add(1) per connection in Serve) or ctx
// then: s.start(); s.stopOnce.Do(func() { close(s.stop) }); <-s.loopDone
```

`s.start()` before stopping makes sure `loopDone` gets closed even if `Serve`
was never called.

A connection accepted just before the listener closed is a sneaky race. Two
rules close it:

- in `Serve`, check `closing` and call `s.handlers.Add(1)` **under `lmu`** –
  then no `Add` can happen after Shutdown has started waiting;
- in `handle`, register the connection only `if !s.closing.Load()` (inside
  `s.call`) – otherwise it would miss the goodbye and Shutdown would wait for
  it forever.

Shutdown may be called more than once; later calls return nil.

## GameTime

```go
total := startMinute + ticks
day = int(total/(24*60)) + 1
hour = int(total/60) % 24
minute = int(total % 60)
```
