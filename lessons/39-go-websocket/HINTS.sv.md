# Tips

## Accept

```go
select {
case c := <-l.conns:
	return c, nil
case <-l.closed:
	return nil, &net.OpError{Op: "accept", Net: "websocket", Addr: l.addr, Err: net.ErrClosed}
}
```

`*net.OpError` har en `Unwrap`-metod, så `errors.Is(err, net.ErrClosed)` hittar det.

## Close – en gång

```go
l.once.Do(func() { close(l.closed) })
return nil
```

Att stänga en stängd kanal ger panik; `sync.Once` gör att andra `Close` inte gör något.

## ServeHTTP

Om `websocket.Accept` misslyckas har den redan skrivit ett felsvar (403 för
ett främmande ursprung) – logga och returnera bara. Om `Offer` returnerar
false håller servern på att stängas: `c.Close(websocket.StatusGoingAway, "...")`.

## main.go

```go
//go:embed web
var webFiles embed.FS

files, _ := fs.Sub(webFiles, "web") // så att "/" är web/index.html
gw := mud.NewGateway(files)
web := &http.Server{Addr: *webAddr, Handler: gw, ReadHeaderTimeout: 10 * time.Second}
wl, err := net.Listen("tcp", *webAddr)
...
go web.Serve(wl)
go srv.Serve(gw.Listener())
```

`ReadHeaderTimeout` stoppar klienter som öppnar en anslutning och aldrig
skickar klart sin förfrågan. Och sätt `OnCopyover` till nil medan webben är på.

## Att testa för hand

Öppna sidan, tryck F12 och titta i konsolen och nätverksfliken:
förfrågan till `/ws` ska säga `101 Switching Protocols`.
