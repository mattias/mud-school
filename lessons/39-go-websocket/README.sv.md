# Lektion 39 – WebSocket

Telnet och din Rust-klient duger för dig, men de flesta kommer bara att
prova din MUD om den ligger ett klick bort: i webbläsaren. Webbläsare kan
inte öppna en vanlig TCP-anslutning – men de kan öppna en **WebSocket**, en
anslutning som börjar som en HTTP-förfrågan och sedan står öppen åt båda hållen.

I den här lektionen får servern en **gateway**: den serverar en webbsida
(given: `web/index.html`, `web/mud.js`, `web/style.css`) och gör varje
WebSocket till en `net.Conn`. Och sedan tricket: gatewayen har en egen
*lyssnare*, och `Serve` – samma `Serve` som för telnet – tar emot spelare från den.

```
webbläsare ──HTTP──▶ Gateway ──filer──▶ index.html, mud.js
webbläsare ──/ws───▶ Gateway ──net.Conn──▶ ConnListener ──Accept──▶ Serve ──▶ handle
telnet     ─────────────────────────────▶ net.Listener ──Accept──▶ Serve ──▶ handle
```

Logga in, gå, prata – resten av servern vet aldrig vem som använder en webbläsare.

## Mål

- Servera filer och en WebSocket med `net/http`.
- Gör en typ som **uppfyller ett gränssnitt** servern redan använder:
  `net.Listener` – med kanaler i stället för en socket.
- Låt resten av programmet vara som det är: **anpassa** nya saker till gamla gränssnitt.
- Förstå varför WebSocketen kontrollerar var sidan kommer ifrån (`Origin`).

## Go-koncept

### En `http.Handler`

```go
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/ws" {
		g.files.ServeHTTP(w, r) // http.FileServerFS
		return
	}
	...
}
```

Allt som har en `ServeHTTP`-metod är en handler: `http.ListenAndServe(addr, g)`.

### En lyssnare av kanaler

`net.Listener` är tre metoder: `Accept`, `Close`, `Addr`. En `ConnListener`
får sina anslutningar från en obuffrad kanal:

```go
func (l *ConnListener) Offer(c net.Conn) bool {
	select {
	case l.conns <- c:   // någon väntar i Accept – lämna över
		return true
	case <-l.closed:     // stängd: ingen kommer någonsin att ta den
		return false
	}
}
```

`Accept` gör en `select` åt andra hållet. `Close` stänger `closed` – en gång,
med `sync.Once` – vilket väcker varje `Accept` och `Offer` som väntar. Därefter
returnerar `Accept` ett fel som `errors.Is(err, net.ErrClosed)`, precis som en
TCP-lyssnare, så `Serve` slutar på samma sätt som alltid.

### Från WebSocket till `net.Conn`

Biblioteket `github.com/coder/websocket` sköter WebSocket-protokollet:

```go
c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: g.Origins})
conn := websocket.NetConn(context.Background(), c, websocket.MessageText)
```

`NetConn` slår in den i en `net.Conn`: det servern skriver blir
textmeddelanden, det webbläsaren skickar kan läsas som en ström. Anslutningen
lever längre än HTTP-förfrågan, så den får `context.Background()` – inte
`r.Context()`, som tar slut när `ServeHTTP` returnerar.

### Origin

En sida på en *annan* webbplats kan också öppna en WebSocket mot din server –
med besökarens webbläsare. `Accept` vägrar det om inte sidan kommer från
samma värd (eller någon av `Origins`). Telnet har inte det problemet; webben har det.

## Kontraktet

- `NewConnListener(addr) *ConnListener` med `Offer(c) bool`, `Accept`,
  `Close` (säker att anropa två gånger) och `Addr`. `Accept` efter `Close`:
  ett fel som är `net.ErrClosed`; `Offer` efter `Close`: false.
- `NewGateway(files fs.FS) *Gateway`, `Listener() net.Listener`, och
  `ServeHTTP`: `/ws` är spelet (textmeddelanden, samma rader som i telnet –
  JSON efter `@hello`), allt annat är filerna. Andra webbplatsers sidor nekas.
- En webbspelares `RemoteAddr` är webbläsarens adress.
- En läs-tidsgräns fungerar som för TCP: läsningen avbryts med
  `os.ErrDeadlineExceeded` och anslutningen står kvar, så att `Shutdown`s hej då
  ändå kommer fram. (`websocket.NetConn` ensam stänger anslutningen i stället –
  slå in den i `newDeadlineConn`, given.)
- `Shutdown` säger hej då till webbspelare också, och `Serve(gw.Listener())`
  returnerar `ErrServerClosed`.
- `main.go`: `-web` (standard `127.0.0.1:8080`, tomt = av) serverar gatewayen,
  med filerna från `web/` inbakade (`//go:embed web`). Medan den är på är
  `@copyover` av (en copyover kan inte lämna över WebSocketar).

## Uppgift

1. `go get github.com/coder/websocket@v1.8.12` i `server/`.
2. `mud/web.go` är delvis given (typerna, `NewGateway`, `newDeadlineConn`):
   skriv `Offer`, `Accept`, `Close` och `ServeHTTP`.
3. `main.go`: flaggan `-web`, `//go:embed web`, en `http.Server` med
   gatewayen, `go srv.Serve(gw.Listener())`, och `web.Shutdown` efter `srv.Shutdown`.

Kör sedan `./mudschool run server` och öppna http://127.0.0.1:8080. Än så
länge spelar du med bara text – kartan kommer i lektion 41.

## Fundera på

- Varför måste `Offer` vänta på `Accept` (en obuffrad kanal)? Vad skulle en
  buffrad kanal ändra vid en avstängning?
- `NetConn` gör meddelanden till en ström. Vad händer om webbläsaren skickar
  `"lo"` och sedan `"ok\n"`?
- Varför kan inte en copyover lämna över en WebSocket, när den kan lämna över
  en TCP-anslutning?
- Vad gör `http.FileServerFS` med en förfrågan efter `/../../etc/passwd`?

## Utmaningar

- Skicka `language` från webbläsarens språk direkt efter inloggningen (lektion 38).
- `ws://` utan webbläsare: anslut med `websocat` och spela från en terminal.
- Få copyover att fungera med webben: lämna över HTTP-lyssnaren som en fil
  också, och låt webbklienterna ansluta igen av sig själva.
