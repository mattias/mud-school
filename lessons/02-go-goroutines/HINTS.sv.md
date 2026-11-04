# Ledtrådar

## Ett ord

Sätt `go` framför anropet till `s.handle(conn)` i `Serve`.

## Loggning

```go
log.Printf("ny anslutning från %s", conn.RemoteAddr())
defer log.Printf("%s kopplade ner", conn.RemoteAddr())
```

Två `defer` i samma funktion körs i omvänd ordning (sist in, först ut).
