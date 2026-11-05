# Ledtrådar

## Loopen

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

## handle med call

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

Defer körs i omvänd ordning: spelaren tas bort ur `s.players` och `s.conns`
*innan* `p.close()` – så ingen kan skicka till en stängd utkorg.

## Shutdown

```go
s.lmu.Lock()                 // den lilla mutexen som skyddar s.listeners
s.closing.Store(true)        // atomic.Bool – Serve kollar den när Accept misslyckas
for _, l := range s.listeners { l.Close() }
s.lmu.Unlock()
s.call(func() {
    for p := range s.conns {
        p.Send(MsgShutdown)
        p.conn.SetReadDeadline(time.Now())
    }
})
// vänta på s.handlers (en sync.WaitGroup, Add(1) per anslutning i Serve) eller ctx
// sedan: s.start(); s.stopOnce.Do(func() { close(s.stop) }); <-s.loopDone
```

`s.start()` innan du stoppar gör att `loopDone` stängs även om `Serve` aldrig
anropats.

En anslutning som accepterades precis innan lyssnaren stängdes är ett lömskt
race. Två regler täpper till det:

- i `Serve`, kolla `closing` och anropa `s.handlers.Add(1)` **under `lmu`** –
  då kan inget `Add` ske efter att Shutdown har börjat vänta;
- i `handle`, registrera anslutningen bara `if !s.closing.Load()` (inne i
  `s.call`) – annars missar den avskedet och Shutdown väntar på den för evigt.

Shutdown får anropas mer än en gång; senare anrop returnerar nil.

## GameTime

```go
total := startMinute + ticks
day = int(total/(24*60)) + 1
hour = int(total/60) % 24
minute = int(total % 60)
```
