# Lesson 39 – WebSockets

Telnet and your Rust client are fine for you, but most people will only ever
try your MUD if it is one click away: in the browser. Browsers can't open a
plain TCP connection – but they can open a **WebSocket**, a connection that
starts as an HTTP request and then stays open in both directions.

In this lesson the server gets a **gateway**: it serves a web page (given:
`web/index.html`, `web/mud.js`, `web/style.css`) and turns every WebSocket
into a `net.Conn`. And then the trick: the gateway has a *listener* of its
own, and `Serve` – the same `Serve` as for telnet – accepts players from it.

```
browser ──HTTP──▶ Gateway ──files──▶ index.html, mud.js
browser ──/ws───▶ Gateway ──net.Conn──▶ ConnListener ──Accept──▶ Serve ──▶ handle
telnet  ─────────────────────────────▶ net.Listener ──Accept──▶ Serve ──▶ handle
```

Log in, walk, talk – the rest of the server never knows who uses a browser.

## Goals

- Serve files and a WebSocket with `net/http`.
- Make a type that **satisfies an interface** the server already uses:
  `net.Listener` – with channels instead of a socket.
- Keep the rest of the program unchanged: **adapt** new things to old interfaces.
- Know why the WebSocket checks where the page comes from (`Origin`).

## Go concepts

### An `http.Handler`

```go
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/ws" {
		g.files.ServeHTTP(w, r) // http.FileServerFS
		return
	}
	...
}
```

Anything with a `ServeHTTP` method is a handler: `http.ListenAndServe(addr, g)`.

### A listener made of channels

`net.Listener` is three methods: `Accept`, `Close`, `Addr`. A `ConnListener`
gets its connections from an unbuffered channel:

```go
func (l *ConnListener) Offer(c net.Conn) bool {
	select {
	case l.conns <- c:   // somebody is in Accept – hand it over
		return true
	case <-l.closed:     // closed: nobody will ever take it
		return false
	}
}
```

`Accept` selects the other way round. `Close` closes `closed` – once, with
`sync.Once` – which wakes every `Accept` and `Offer` that waits. After that,
`Accept` returns an error that `errors.Is(err, net.ErrClosed)`, just like a
TCP listener, so `Serve` stops the way it always has.

### From WebSocket to `net.Conn`

The library `github.com/coder/websocket` does the WebSocket protocol:

```go
c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: g.Origins})
conn := websocket.NetConn(context.Background(), c, websocket.MessageText)
```

`NetConn` wraps it in a `net.Conn`: what the server writes becomes text
messages, what the browser sends can be read as a stream. The connection
outlives the HTTP request, so it gets `context.Background()` – not
`r.Context()`, which ends when `ServeHTTP` returns.

### Origin

A page on *another* site can also open a WebSocket to your server – with
the visitor's browser. `Accept` refuses that unless the page comes from the
same host (or one of `Origins`). Telnet has no such problem; the web has.

## The contract

- `NewConnListener(addr) *ConnListener` with `Offer(c) bool`, `Accept`,
  `Close` (safe twice) and `Addr`. `Accept` after `Close`: an error that is
  `net.ErrClosed`; `Offer` after `Close`: false.
- `NewGateway(files fs.FS) *Gateway`, `Listener() net.Listener`, and
  `ServeHTTP`: `/ws` is the game (text messages, the same lines as telnet –
  JSON after `@hello`), everything else the files. Other sites' pages are refused.
- A web player's `RemoteAddr` is the browser's address.
- `Shutdown` says goodbye to web players too, and `Serve(gw.Listener())`
  returns `ErrServerClosed`.
- `main.go`: `-web` (default `127.0.0.1:8080`, empty = off) serves the gateway,
  with the files from `web/` baked in (`//go:embed web`). While it is on,
  `@copyover` is off (a copyover can't hand over WebSockets).

## Task

1. `go get github.com/coder/websocket@v1.8.12` in `server/`.
2. `mud/web.go` is partly given: write `Offer`, `Accept`, `Close` and `ServeHTTP`.
3. `main.go`: the `-web` flag, `//go:embed web`, an `http.Server` with the
   gateway, `go srv.Serve(gw.Listener())`, and `web.Shutdown` after `srv.Shutdown`.

Then `./mudschool run server` and open http://127.0.0.1:8080. You play with
text only for now – the map comes in lesson 41.

## Think about

- Why must `Offer` wait for `Accept` (an unbuffered channel)? What would a
  buffered channel change during a shutdown?
- `NetConn` turns messages into a stream. What happens if the browser sends
  `"lo"` and then `"ok\n"`?
- Why can't a copyover hand over a WebSocket, when it can hand over a TCP connection?
- What does `http.FileServerFS` do with a request for `/../../etc/passwd`?

## Challenges

- Send `language` from the browser's language right after login (lesson 38).
- `ws://` without a browser: connect with `websocat` and play from a terminal.
- Make copyover work with the web: hand over the HTTP listener as a file too,
  and let the web clients reconnect by themselves.
