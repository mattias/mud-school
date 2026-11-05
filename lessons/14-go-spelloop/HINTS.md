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
s.closing.Store(true)        // atomic.Bool – Serve checks it after Accept fails
// close every listener (keep them in a slice protected by a small mutex)
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

## GameTime

```go
total := startMinute + ticks
day = int(total/(24*60)) + 1
hour = int(total/60) % 24
minute = int(total % 60)
```
