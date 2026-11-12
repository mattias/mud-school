# Hints

## Accept

```go
select {
case c := <-l.conns:
	return c, nil
case <-l.closed:
	return nil, &net.OpError{Op: "accept", Net: "websocket", Addr: l.addr, Err: net.ErrClosed}
}
```

`*net.OpError` has an `Unwrap` method, so `errors.Is(err, net.ErrClosed)` finds it.

## Close – once

```go
l.once.Do(func() { close(l.closed) })
return nil
```

Closing a closed channel panics; `sync.Once` makes the second `Close` a no-op.

## ServeHTTP

If `websocket.Accept` fails it has already written an error response
(403 for a foreign origin) – just log and return. If `Offer` returns false
the server is shutting down: `c.Close(websocket.StatusGoingAway, "...")`.

## main.go

```go
//go:embed web
var webFiles embed.FS

files, _ := fs.Sub(webFiles, "web") // so that "/" is web/index.html
gw := mud.NewGateway(files)
web := &http.Server{Addr: *webAddr, Handler: gw, ReadHeaderTimeout: 10 * time.Second}
wl, err := net.Listen("tcp", *webAddr)
...
go web.Serve(wl)
go srv.Serve(gw.Listener())
```

`ReadHeaderTimeout` stops clients that open a connection and never finish
their request. And set `OnCopyover` to nil while the web is on.

## Testing by hand

Open the page, press F12 and look at the console and the network tab: the
`/ws` request should say `101 Switching Protocols`.
