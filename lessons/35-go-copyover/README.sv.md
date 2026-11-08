# Lektion 35 – Copyover

Data, skript och plugins kan laddas om i drift. Men vissa ändringar kräver
ny Go-kod – en ny funktion i spelloopen, en buggfix i striden. Det betyder ett
nytt program, och ett nytt program betyder en omstart. Eller?

**Copyover** är ett trick som MUD:ar har använt sedan 90-talet. Den körande
servern startar den **nya** versionen av sig själv och lämnar över allt som
måste överleva: den lyssnande socketen, varje spelares öppna anslutning och
spelarnas tillstånd. Sedan avslutas den gamla processen. Spelarna ser en rad:

```
Världen skälver ett ögonblick ... allt är som förut, och ändå nytt.
```

– och spelar vidare, på ny kod, utan att någonsin ha kopplats ner.

```
gammal process                                ny process (det nya programmet)
──────────────                                ───────────────────────────────
duplicera lyssnaren och varje anslutning ──▶  fd 3 = lyssnaren, fd 4, 5, ... = spelarna
skriv tillståndet (rum, HP, saker) som JSON ─▶ MUD_COPYOVER=/tmp/state.json
starta det nya programmet (ExtraFiles)
stäng sina egna kopior, avsluta tyst           net.FileListener(fd 3), net.FileConn(fd 4) ...
                                               adoptera spelarna, Serve på lyssnaren
```

## Mål

- Förstå **fildeskriptorer**: en anslutning är ett nummer som operativsystemet
  ger en process – och det numret kan lämnas vidare till en barnprocess.
- Använda `exec.Cmd.ExtraFiles`, `(*net.TCPConn).File()`, `net.FileConn` och `net.FileListener`.
- Dela ett ansvar i två: "logga in" och "spela" (`handle` → `play`), så att en
  spelare kan börja på "spela".
- Lämna över tillstånd mellan processer – serialiserat som JSON.

## Go-koncept

### Fildeskriptorer

`conn.(*net.TCPConn).File()` returnerar en `*os.File` med en **dubblett** av
anslutningens deskriptor – ett andra nummer för samma socket. En socket
stängs först när *alla* deskriptorer för den är stängda. Alltså:

1. den gamla processen duplicerar (`File()`),
2. lämnar dubbletterna till den nya processen (`cmd.ExtraFiles` – de blir
   fd 3, 4, 5 … där),
3. stänger sina egna deskriptorer – anslutningen förblir öppen, eftersom den
   nya processen har kopior,
4. den nya processen gör om numren till Go-värden:
   `net.FileConn(os.NewFile(4, "anna"))`.

### Två sorters farväl

`Shutdown` (lektion 14) säger adjö och loggar ut alla. `HandOver` får göra
ingetdera: inget meddelande, ingen sparning, inget "Anna gick". En flagga,
`s.handingOver`, säger till spelarnas goroutiner att hoppa över sin städning.

### Den riktiga exec:en

`main.go` gör operativsystemdelen (given): `startNewServer` skriver
tillståndet till en temporär fil och startar `os.Executable()` – programmet
på disken, så ett nybyggt program betyder ny kod – med `ExtraFiles` och
miljövariabeln `MUD_COPYOVER`. När `main` startar med den variabeln satt
adopterar den filerna i stället för att öppna en ny lyssnare.

### Windows

Windows har inget sätt att lämna en socket till en ny process på det här
viset: `(*net.TCPConn).File()` ger ett fel där, och `@copyover` säger det.
Prova copyover i WSL. Testerna som behöver det hoppas över på Windows.

## Kontraktet

- `CopyoverPlayer {Name, Room, X, Y, HP, Items, Structured, File}` och
  `CopyoverState {Ticks, Players}` med JSON-namnen `name`, `room`, `x`, `y`,
  `hp`, `items`, `structured`, `file`, `ticks`, `players`.
- `(s *Server) PrepareCopyover() (*CopyoverState, []*os.File, error)` –
  anropas utanför spelloopen; `files[0]` är lyssnaren, varje spelares `File`
  är indexet för anslutningens fil.
- `(s *Server) HandOver(ctx) error` – stoppar den här servern utan adjö, utan
  att logga ut någon och utan att spara; anslutningarna förblir öppna för den nya processen.
- `(s *Server) Adopt(state, files) (net.Listener, error)` – den nya servern
  tar tillbaka spelarna: samma rum (eller startrummet), position (om den
  fortfarande är gångbar), HP och saker; var och en får `MsgCopyoverDone` och
  en `look`, och spelar vidare utan att logga in. Den returnerar lyssnaren till `Serve`.
- `(s *Server) Done() <-chan struct{}` – stängs när servern har stannat.
- `@copyover` (administratörer): med `Options.OnCopyover` satt, förbered →
  starta den nya servern → lämna över; vid ett fel får administratören `MsgCopyoverFailed`.
- `handle` delas upp: inloggningen i `handle`, allt därefter i `play(p)`.

## Uppgift

1. Dela `handle` i `handle` och `play`, och låt `play`:s städning göra
   ingenting medan `s.handingOver` är satt.
2. `copyover.go` är delvis given (typerna, `fileOf`, `Done`, `cmdCopyover`,
   `copyover`): skriv `PrepareCopyover`, `HandOver`, `Adopt` och `adopt`.
3. `Options.OnCopyover`, fälten och kommandot; `main.go` är given.

Prova på riktigt (Linux, macOS eller WSL):

```
cd workspace/server
go build -o mudserver . && ./mudserver -admin Anna
# logga in som Anna i en annan terminal, ändra ett meddelande i messages.go, sedan:
go build -o mudserver .
# och i spelet: @copyover
```

(Använd ett byggt program, inte `go run`: `go run` bygger ett temporärt
program och raderar det – den nya processen skulle starta den *gamla* koden.)

## Fundera på

- Vad händer med ett kommando en spelare skriver *under* copyover?
- Vilket tillstånd lämnas *inte* över (strider, utforskade zoner, varelsernas
  positioner, …)? Vad av det skulle spelarna märka?
- Tänk om det nya programmet kraschar när det startar? Hur skulle det gamla
  kunna märka det och fortsätta?
- Varför `os.Executable()` och inte `os.Args[0]`?

## Utmaningar

- Lämna över mer: strider, varelsernas positioner, klockan.
- En hälsokontroll: den nya processen säger "jag är uppe" på ett rör innan den
  gamla avslutas; annars fortsätter den gamla.
- Kodbyte i drift på Erlangs vis är något helt annat – läs om det och jämför.
