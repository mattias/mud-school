# Hints

## Allow

```go
if !r.last.IsZero() {
	r.tokens += now.Sub(r.last).Seconds() * r.rate
	if r.tokens > r.burst {
		r.tokens = r.burst
	}
}
r.last = now
if r.tokens < 1 {
	return false
}
r.tokens--
return true
```

## allowCommand

```go
if p.limiter == nil || p.limiter.Allow(time.Now()) {
	p.dropped = 0
	metricCommands.Add(1)
	return true, false
}
```

…and otherwise count, `p.dropped++`, `MsgSlowDown` when `p.dropped == 1`,
`MsgFlooding` and `quit = true` at `FloodLimit`.

## play

```go
if ok, quit := s.allowCommand(p); quit {
	return
} else if !ok {
	continue // dropped
}
```

## handle – refusing a connection

Decide inside the `s.call` that registers the connection (only the game loop
may read `s.conns`), and send `MsgTooManyConnections` after it – `p.close()`
(deferred) makes sure it is sent before the connection closes.

## The login deadline

```go
if !ok {
	if errors.Is(p.scanner.Err(), os.ErrDeadlineExceeded) && !s.closing.Load() {
		p.SendKind(KindError, MsgLoginTimeout)
	}
	return
}
```

and in the game loop, after a successful login:
`if !s.closing.Load() { conn.SetReadDeadline(time.Time{}) }`.

## The debug server

```go
import _ "net/http/pprof"
...
go func() { log.Print(http.ListenAndServe(cfg.Debug, nil)) }() // nil = DefaultServeMux
```
