# Lektion 43 – Skydd

Din MUD ligger på internet. Förr eller senare klistrar någon in en hel bok i
den, skriptar tusen kommandon i sekunden, öppnar hundra anslutningar eller
ansluter och aldrig säger ett ord. Inget av det får göra spelet långsamt för
alla andra – kom ihåg att varje kommando körs i **en** spelloop.

```
> say spam
> say spam
... (500 till)
Ta det lugnt – du skickar kommandon för fort.
...
För många kommandon – hej då.
```

## Mål

- Begränsa en takt med en **token bucket** – och testa den utan att vänta.
- Lägg gränserna där de är billiga: *innan* ett kommando når spelloopen.
- Använd tidsgränser mot klienter som aldrig blir klara (`SetReadDeadline`).
- Räkna det som händer med `expvar`, och se vart tiden tar vägen med `pprof`.

## Go-koncept

### En token bucket

```
burst ┃■■■■■┃  varje kommando tar en token;
      ┃■■■  ┃  `rate` tokens rinner tillbaka per sekund;
      ┃     ┃  ingen token – kommandot kastas.
```

```go
func (r *RateLimiter) Allow(now time.Time) bool
```

Korta skurar går bra (fem snabba steg), en jämn flod gör det inte. Tiden är
en parameter, så testet kan säga "en halv sekund senare" utan att sova.
Begränsaren lever i spelarens egen goroutine (`play`) – ett kastat kommando
når aldrig spelloopen alls.

### Svara tyst på en flod

En flod på 500 rader får inte 500 svar – det skulle fördubbla den. Spelaren
får `MsgSlowDown` en gång per flod, och kopplas ner efter `FloodLimit`
kastade kommandon i rad.

### Gränser för läsning

```go
p.scanner.Buffer(make([]byte, 0, min(1024, max)), max)   // längre rader: ErrTooLong
conn.SetReadDeadline(time.Now().Add(s.loginTimeout))        // logga in i tid
```

`bufio.Scanner` vägrar en rad som är längre än buffertens gräns; det avslutar
anslutningen med `MsgLineTooLong`. (Gränsen är den största av `max` och
buffertens kapacitet – börja litet.) Läs-tidsgränsen från lektion 14:s
avstängning återanvänds: inloggningen måste ske i tid, och tidsgränsen tas
bort när den har skett – om inte en avstängning har satt en under tiden.

### Per adress

`p.ip` är värddelen av `RemoteAddr` – bakom Caddy, tack vare `TrustProxy`
(lektion 42), spelarens riktiga adress. Spelloopen räknar anslutningarna i
`s.conns` med samma `ip` innan den släpper in en ny.

### `expvar` och `pprof`

```go
var metricCommands = expvar.NewInt("mud_commands_total")
```

`expvar` publicerar räknare som JSON på `/debug/vars`; `net/http/pprof` lägger
till `/debug/pprof/`. Båda registrerar sig på `http.DefaultServeMux`, som
servern serverar på en egen adress, `-debug 127.0.0.1:6060` – aldrig på den publika:

```
go tool pprof http://127.0.0.1:6060/debug/pprof/profile?seconds=10
```

visar vilka funktioner processorn är upptagen med. Lektion 44 ger den något att göra.

## Kontraktet

- `Options.MaxPerIP`, `CommandRate`, `CommandBurst`, `MaxLineLength`
  (0 = `DefaultMaxLine`), `LoginTimeout`. Noll betyder ingen gräns.
- `NewRateLimiter(rate, burst)` börjar full; `Allow(now)` fyller på med tiden
  sedan förra anropet, aldrig över `burst`.
- För många anslutningar från en IP: `MsgTooManyConnections`, stängd – och
  det finns plats igen när en stängs.
- Kommandon över takten kastas; `MsgSlowDown` en gång per flod;
  `MsgFlooding` och nedkopplad efter `FloodLimit` i rad.
- En rad längre än gränsen: `MsgLineTooLong`, stängd. Inte inloggad inom
  `LoginTimeout`: `MsgLoginTimeout`, stängd. Inloggad: ingen tidsgräns.
- `expvar`: `mud_connections_total`, `mud_connections_refused_total`,
  `mud_commands_total`, `mud_commands_dropped_total`, `mud_players`.
- `main`: `-max-per-ip` (5), `-rate` (10), `-burst` (20), `-login-timeout`
  (1m) och `-debug` (av), med `MUD_MAX_PER_IP`, `MUD_RATE`, `MUD_BURST`,
  `MUD_LOGIN_TIMEOUT` och `MUD_DEBUG`.

## Uppgift

1. `mud/protect.go` är delvis given (konstanterna, räknarna, `hostOf`):
   skriv `Allow`, `connsFrom` och `allowCommand`.
2. `server.go`: alternativen och fälten; i `handle` IP-adressen, scannerns
   gräns, kontrollen per IP och inloggningens tidsgräns – båda inne i den
   `s.call` som registrerar anslutningen, så att en avstängning inte kan
   smita in emellan – och att den tas bort efter inloggningen; i `play`
   begränsaren. Räkna spelare i `addPlayer`/`removePlayer`.
   `copyover.go` (lektion 35): `adopt` sätter `p.ip` och scannerns gräns, och
   räknar spelaren också.
3. `player.go`: `ip`, `limiter`, `dropped`, och `MsgLineTooLong` i `readLine`.
4. `config.go` och `main.go`: de nya flaggorna, och debug-servern.

## Fundera på

- Varför ligger takt-begränsaren i spelarens goroutine, men räkningen per IP
  i spelloopen?
- En skola eller ett kafé delar en IP-adress. Vad betyder `-max-per-ip 5` för dem?
- Bcrypt (lektion 20) är långsam med flit. Vad skulle en angripare kunna göra
  med tusen inloggningar på en gång? Vad skulle du begränsa?
- Vilka siffror skulle du vilja ha på en instrumentpanel?

## Utmaningar

- En *global* gräns: högst N kommandon per tick i spelloopen, rättvist fördelade.
- Stäng ute en adress i tio minuter efter en flod.
- Exportera räknarna i Prometheus textformat på `/metrics`.
