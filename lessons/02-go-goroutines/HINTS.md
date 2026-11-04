# Hints

## One word

Put `go` in front of the call to `s.handle(conn)` in `Serve`.

## Logging

```go
log.Printf("new connection from %s", conn.RemoteAddr())
defer log.Printf("%s disconnected", conn.RemoteAddr())
```

Two `defer`s in the same function run in reverse order (last in, first out).
