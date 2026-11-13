# Tips

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

…och annars: räkna, `p.dropped++`, `MsgSlowDown` när `p.dropped == 1`,
`MsgFlooding` och `quit = true` vid `FloodLimit`.

## play

```go
if ok, quit := s.allowCommand(p); quit {
	return
} else if !ok {
	continue // kastat
}
```

## handle – att neka en anslutning

Bestäm inne i `s.call` som registrerar anslutningen (bara spelloopen får läsa
`s.conns`), och skicka `MsgTooManyConnections` efter den – `p.close()`
(med defer) ser till att det skickas innan anslutningen stängs.

## Inloggningens tidsgräns

Sätt den i samma `s.call` som lägger till `p` i `s.conns`, precis innan: om
den sattes efteråt, utanför loopen, skulle en `Shutdown` däremellan sätta sin
egen tidsgräns (nu) först – och din skulle skriva över den med en minut.
Efter en misslyckad inloggning:

```go
if !ok {
	if errors.Is(p.scanner.Err(), os.ErrDeadlineExceeded) && !s.closing.Load() {
		p.SendKind(KindError, MsgLoginTimeout)
	}
	return
}
```

och i spelloopen, efter en lyckad inloggning:
`if !s.closing.Load() { conn.SetReadDeadline(time.Time{}) }`.

## Debug-servern

```go
import _ "net/http/pprof"
...
go func() { log.Print(http.ListenAndServe(cfg.Debug, nil)) }() // nil = DefaultServeMux
```
