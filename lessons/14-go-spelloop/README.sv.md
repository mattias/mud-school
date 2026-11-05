# Lektion 14 – Spelloopen

Fas 3 börjar: världen ska leva även när ingen skriver något. Klockan tickar,
varelser vandrar (lektion 16), strider går runda för runda (lektion 18). Det
kräver en **spelloop** – och det är det perfekta tillfället att tänka om hur
servern skyddar sin data.

Hittills har varje goroutine låst `s.mu` när den rört världen. Det fungerar,
men blir svårare med varje funktion: vilka funktioner håller låset? Vad händer
om ett tick och ett kommando körs samtidigt? Många spelservrar (och många
Go-program) använder en annan modell:

> **En goroutine äger allt tillstånd.** Alla andra skickar *arbete* till den
> på en kanal och väntar, om de behöver ett svar. Det är *actor*-mönstret –
> "kommunicera inte genom att dela minne; dela minne genom att kommunicera".

```
 anslutningarnas goroutines         spelloopen (en goroutine)
 ──────────────────────────         ──────────────────────────
 readLine ── s.call(dispatch) ──►   select {
 readLine ── s.call(dispatch) ──►     case f := <-s.events: f()
                                      case <-ticker.C:      s.ticks++; s.onTick()
 Shutdown ── close(s.stop) ─────►     case <-s.stop:        return
                                    }
```

Från och med den här lektionen blir stöttningen tunnare: du får ett kontrakt
och tester, och designen är din.

## Mål

- Strukturera om ett program kring en goroutine som äger tillståndet.
- Använda `select` över flera kanaler, och `time.Ticker`.
- Använda `context.Context` för en tidsgräns och stänga av en server snyggt.
- Konfigurera med en options-struct.

## Go-koncept

### Funktioner på en kanal

```go
events := make(chan func())
go func() {
    for f := range events {
        f()             // körs i den här goroutinen, en i taget
    }
}()

done := make(chan struct{})
events <- func() { counter++; close(done) }
<-done              // vänta på svaret
```

`do` (skicka och gå vidare) och `call` (vänta tills det är gjort) i `loop.go`
är precis det, plus en utväg om loopen har stannat.

> **Anropa aldrig `s.call` inifrån spelloopen.** Loopen skulle vänta på sig
> själv – en deadlock. Kommandon körs redan i loopen, så de använder världen
> direkt.

### context

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
select {
case <-done:          // klart i tid
case <-ctx.Done():    // gav upp: ctx.Err() är context.DeadlineExceeded
}
```

## Kontraktet

```go
type Options struct {
    World *World        // nil = DefaultWorld()
    Tick  time.Duration // 0 = DefaultTick (100 ms)
}
func NewServerWithOptions(o Options) *Server  // NewServer/NewServerWithWorld använder den
func (s *Server) Ticks() int64                 // hur många ticks som körts
func (s *Server) Shutdown(ctx context.Context) error
var ErrServerClosed                            // det Serve returnerar efter Shutdown
func GameTime(ticks int64) (day, hour, minute int) // ett tick = en spelminut, från dag 1 08:00
```

och kommandot `time`, som skickar `MsgTime`.

## Uppgift

1. **`loop.go`** (ny): skriv `run`, `Ticks` och `Shutdown`. `do` och `call` är givna.
2. **`clock.go`** (ny): skriv `GameTime` och registrera `time` i `commands.go`.
3. **`server.go`**:
   - lägg till `Options`, `DefaultTick`, `ErrServerClosed`, `NewServerWithOptions`
     och loopens fält (se kommentaren i `loop.go`), samt
     `conns map[*Player]bool` för alla anslutningar och `ticks int64`,
   - **ta bort `mu`**: `players`, `conns`, `ticks` och `p.Room` ägs nu av spelloopen,
   - i `handle` och `login`: kör allt som rör världen med `s.call`: att
     registrera anslutningen, `addPlayer`, varje kommando
     (`s.call(func() { s.dispatch(p, line) })`) och städningen när spelaren
     går. Använd `call` (inte `do`) så att kommandon för en spelare körs i
     ordning och `p.quit` är satt innan du tittar på den,
   - `Serve` startar loopen, kommer ihåg sin lyssnare (Shutdown måste stänga
     den) och returnerar `ErrServerClosed` efter Shutdown. Shutdown får anropas mer än en gång; senare anrop returnerar nil.
4. **`rooms.go`/`commands.go`**: ta bort låsen. Hjälpfunktionerna `…Locked`
   betyder nu helt enkelt "körs i spelloopen" – döp om dem (`look`,
   `describeRoom`, `playersInRoom`, `broadcastRoom`, …). `who` använder den
   loop-interna `playerNames`; den exporterade `PlayerNames` använder `s.call`.
5. **`main.go`** (frivilligt men trevligt): `signal.NotifyContext` så att
   Ctrl+C anropar `Shutdown` och alla spelare får ett hejdå. Lägg till en
   `-tick`-flagga.

Kör `./mudschool check --race` minst en gång – race detectorn är ditt bevis
på att bara loopen rör världen.

## Testa själv

```
./mudschool run server -tick 50ms
```

Logga in, skriv `time` några gånger. Tryck sedan Ctrl+C i serverfönstret och
se klienterna få `MsgShutdown`.

## Fundera på

- Vad skulle gå fel om `handle` använde `s.do` istället för `s.call` för kommandon?
- `Shutdown` använder `SetReadDeadline` för att väcka goroutines som väntar i
  `readLine`. Varför kan inte loopen bara stänga spelarens utkorg?
- Mutex eller actor – vilka fördelar har de? När skulle du ändå välja en mutex?

## Utmaningar

- Låt en klocka slå varje spel-timme: `onTick` skickar ett meddelande till alla när minuten är 0.
- Dag och natt: låt `look` lägga till "Det är mörkt." mellan 22:00 och 06:00.
