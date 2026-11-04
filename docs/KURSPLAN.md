# Kursplan – MUD School

> Lär dig **Go** och **Rust** genom att bygga en MUD-server och en matchande
> klient från grunden – från en ekoserver på 20 rader till en utbyggbar
> spelvärld med tile-grafik, skript och plugins som laddas utan omstart.

Det här dokumentet beskriver *hela* resan: varför den ser ut som den gör,
vad du lär dig i varje steg och vart vi är på väg. Lektion 01–13 finns
implementerade i `lessons/` (med tester och referenslösningar). Faserna 3–6
är planerade och beskrivna här så att du ser helheten – och kan bygga vidare
på egen hand.

---

## Innehåll

1. [Pedagogisk idé](#pedagogisk-idé)
2. [Arbetssättet i varje lektion](#arbetssättet-i-varje-lektion)
3. [Översikt](#översikt)
4. [Fas 1 – Grunderna (lektion 01–05)](#fas-1--grunderna-lektion-0105)
5. [Fas 2 – En värld och ett ansikte (lektion 06–13)](#fas-2--en-värld-och-ett-ansikte-lektion-0613)
6. [Fas 3 – En levande värld](#fas-3--en-levande-värld-planerad)
7. [Fas 4 – Tiles och zoner](#fas-4--tiles-och-zoner-planerad)
8. [Fas 5 – En utbyggbar server](#fas-5--en-utbyggbar-server-planerad)
9. [Fas 6 – Ut i världen](#fas-6--ut-i-världen-valfria-spår)
10. [Begreppskarta: Go och Rust sida vid sida](#begreppskarta-go-och-rust-sida-vid-sida)
11. [Läsanvisningar](#läsanvisningar)
12. [Tidsåtgång](#tidsåtgång)

---

## Pedagogisk idé

### 1. Ett spel som växer – inte övningar som försvinner

Du bygger **ett** program genom hela kursen. Koden från lektion 04 lever
kvar i lektion 13. Det betyder att du måste leva med dina egna beslut,
refaktorera när världen förändras – precis som i riktiga projekt. Varje
lektion slutar med något du kan *köra och visa* för någon.

### 2. Koncept precis när problemet kräver dem (just-in-time)

Inget språkkoncept introduceras "för att det står i boken". Varje koncept
dyker upp när spelet behöver det:

| Problemet | Konceptet |
|---|---|
| Den andra spelaren kommer inte in | goroutines (02) |
| Två spelare loggar in samtidigt och servern kraschar | `sync.Mutex`, data races (05) |
| En spelare med dålig uppkoppling fryser hela servern | kanaler, `select`, en skrivare per socket (07) |
| Klienten missar meddelanden medan du skriver | trådar och kanaler i Rust (06) |
| Kartan ska vara grön i en klient och en bild i en annan | gränssnitt, separera betydelse från utseende (10) |
| Servern skickar olika sorters meddelanden | enums med data, `match` (11) |
| 600 rum går inte att skriva i kod | JSON, validering, `go:embed` (09) |

Det gör att du förstår *varför* ett koncept finns – inte bara hur det skrivs.

### 3. Två språk, med avsikt

**Go först.** Go är litet och förlåtande, och nätverksprogrammering är
dess hemmaplan. Du har en fungerande flerspelarserver efter två lektioner.
Den tidiga framgången bär dig genom de svårare delarna.

**Rust när det finns något att prata med.** Rust introduceras i lektion 03,
när servern redan finns. Rust är svårare, men du lär dig det med en konkret
uppgift och ett facit att testa mot (din egen server).

**Kontrastiv inlärning.** Varje Rust-lektion har en *Go ↔ Rust*-jämförelse.
Att se samma idé (en kanal, ett fel, ett gränssnitt) uttryckas på två sätt
gör att du förstår idén bättre än om du bara kunde ett språk. Du lär dig
skilja på *vad* (samtidighet, felhantering) och *hur* (goroutines vs trådar,
`error` vs `Result`).

**Rätt verktyg för jobbet.** Fördelningen är inte godtycklig: Go för en
nätverksserver med tusentals samtidiga anslutningar, Rust för en klient med
grafik och full kontroll. I fas 5 möts språken: plugins skrivna i Rust,
kompilerade till WebAssembly, körs inuti Go-servern.

### 4. Testerna är specifikationen

Varje lektion har tester som `./mudschool check` kör. De är skrivna för att
vara läsbara: felmeddelandena säger vad som förväntades och vad som hände,
på svenska. Arbetsgången blir röd → grön → refaktorera. `./mudschool watch`
kör testerna varje gång du sparar, så återkopplingen kommer på sekunder.

Många av serverns tester är **svartlådetester över TCP**: de ansluter till
din server precis som en riktig klient. Det betyder att du är fri att
strukturera din kod som du vill – så länge den beter sig rätt.

### 5. Stöttning som avtar (scaffolding → fading)

| Fas | Du får | Du skriver |
|---|---|---|
| 1 | skelett med `TODO` och nästan färdiga funktionssignaturer | funktionskropparna |
| 2 | nya filer med typer och signaturer, beskrivning av resten | funktioner, ändringar i befintlig kod |
| 3–4 | ett kontrakt (protokoll, gränssnitt) och tester | designen, och egna tester utöver kursens |
| 5–6 | ett mål och några riktlinjer | allt, inklusive testerna |

### 6. Hjälp i lagom doser

- `./mudschool hint` ger ledtrådar **en i taget** – från en knuff i rätt
  riktning till nästan-kod. Försök själv först.
- `./mudschool solution` visar referenslösningen. Använd den för att
  *jämföra efteråt* (`diff -ru`), inte för att kopiera.
- **Fundera på** – reflektionsfrågor i varje lektion. Svara på dem (gärna
  skriftligt, eller högt för en studiekamrat). Det är där förståelsen sätter sig.
- **Utmaningar** – frivilliga utbyggnader för dig som vill mer. Testerna
  påverkas inte.

### 7. Repetition med variation

Kärnidéer återkommer i allt svårare former: kanaler i Go (07) → kanaler i
Rust (06, 13) → actor-mönstret för spelloopen (fas 3) → `tokio::select!`
(fas 6). Felhantering: `error` (01) → egna fel (05) → `errors.Join` (09) →
`Result` och `?` (03) → egna feltyper (11). Varje gång du ser en idé igen
förstår du den lite djupare.

---

## Arbetssättet i varje lektion

```
./mudschool show        # 1. läs: mål, koncept, uppgift
                        # 2. skriv kod i workspace/
./mudschool watch       # 3. testerna körs varje gång du sparar
./mudschool hint        # 4. kör du fast? en ledtråd i taget
./mudschool run server  # 5. prova på riktigt (telnet, din klient)
                        # 6. svara på "Fundera på", prova en utmaning
                        # 7. jämför med referenslösningen
```

Varje README följer samma mall: **Mål → koncept med små exempel → Uppgift →
Testa själv → Fundera på → Utmaningar**.

---

## Översikt

```
Fas 1  Grunderna             01 ─ 05   Go: TCP, goroutines, kommandon, spelare
                                       Rust: första klienten
Fas 2  En värld och ett      06 ─ 13   Go: chatt, rum, världsfiler, protokoll + karta
       ansikte                         Rust: trådar, serde, färgkarta, TUI med ratatui
─────────────────────────────────────────────────── implementerat ovanför ───
Fas 3  En levande värld      14 ─ 22   spelloop, föremål, NPC:er, strid, sparande
Fas 4  Tiles och zoner       23 ─ 30   zoner à la Caves of Qud, procedurgenerering,
                                       synfält, A*, grafisk klient med tileset
Fas 5  En utbyggbar server   31 ─ 36   omladdning live, Lua-skript, WASM-plugins
                                       i Rust, hooks, copyover, bygga i spelet
Fas 6  Ut i världen          valfritt  webbklient (Rust→WASM), async Rust,
                                       GMCP, drift och lasttest
```

---

## Fas 1 – Grunderna (lektion 01–05)

**Mål:** en flerspelarserver du kan logga in på med telnet, och en första
klient i Rust. Du kan läsa och skriva enkla program i båda språken.

| # | Spår | Lektion | Spelet får | Språket lär dig |
|---|---|---|---|---|
| 01 | Go | Hej, TCP! | en ekoserver | paket, funktioner, metoder, structs, `error`, `defer`, `net`, `bufio` |
| 02 | Go | Många spelare samtidigt | flera samtidiga spelare | goroutines, closures, `sync.WaitGroup`, race detectorn |
| 03 | Rust | Hej, Rust! | en egen klient | Cargo, moduler, `struct`/`impl`, `Result`/`?`, `Option`, ägarskap och lån, `String`/`&str` |
| 04 | Go | Kommandon | `look`, `help`, förkortningar, kommandoregister | `strings`, maps, slices, funktioner som värden, tabelldrivna tester |
| 05 | Go | Spelare | inloggning, unika namn, `who` | egna fel, `errors.Is`, `%w`, runor och UTF-8, `sync.Mutex` |

## Fas 2 – En värld och ett ansikte (lektion 06–13)

**Mål:** en värld att utforska, ett protokoll för grafiska klienter och en
klient med färgkarta, statusrad och logg i terminalen.

| # | Spår | Lektion | Spelet får | Språket lär dig |
|---|---|---|---|---|
| 06 | Rust | Trådar och kanaler | klienten lyssnar och skriver samtidigt | `thread::spawn`, `move`, `mpsc`, generics, trait bounds, `Send`/`'static` |
| 07 | Go | Chatt | `say`, in/utloggning meddelas, skydd mot långsamma klienter | kanaler, buffrade kanaler, `select`/`default`, mottryck |
| 08 | Go | Rum | rum, utgångar, `look`, förflyttning, `shout` | pekare, variadiska funktioner, closures som fabriker, låskonventioner |
| 09 | Go | Världsfiler | 21 rum i JSON, validering, `-world`-flagga | `encoding/json`, struct-taggar, `io.Reader`, `errors.Join`, `go:embed` |
| 10 | Go | Protokoll och karta | JSON-protokoll, karta, status – och telnet fungerar fortfarande | gränssnitt/kodningar, `sync/atomic`, 2D-rutnät |
| 11 | Rust | Protokoll i Rust | klienten förstår JSON | enums med data, `match`, `let-else`, serde, `Display` |
| 12 | Rust | Kartan i färg | färgade tiles i Dwarf Fortress-stil | `Copy`, iteratorer, `fmt::Write`, ANSI-koder |
| 13 | Rust | Ett riktigt gränssnitt | TUI: statusrad, logg, karta, inmatning | ratatui, `From`/`Into`, händelse-enum, `TestBackend` |

**Efter fas 2** har du en riktig liten MUD. Det här är en bra punkt att
stanna upp: visa någon, skriv om en del du inte är nöjd med, gör några
utmaningar.

---

## Fas 3 – En levande värld (planerad)

**Mål:** världen lever även när ingen skriver något. Föremål, varelser,
strid – och spelare som finns kvar efter en omstart.

| # | Spår | Lektion | Innehåll |
|---|---|---|---|
| 14 | Go | Spelloopen | En `time.Ticker` driver världen (t.ex. 10 tick/s). Refaktorera från mutex till **actor-mönstret**: *en* goroutine äger hela världen och tar emot kommandon på en kanal. `context.Context` för snygg nedstängning. Jämför de två modellerna. |
| 15 | Go | Föremål och inventarier | `get`, `drop`, `inventory`, `examine`. Gränssnitt (`Describer`), inbäddning av structs, generiska hjälpfunktioner (`func Find[T Named](items []T, name string) T`). |
| 16 | Go | Varelser | NPC:er som vandrar, pratar och reagerar. Tillståndsmaskiner, ett `Behavior`-gränssnitt, `math/rand/v2` med fast seed för testbarhet. |
| 17 | Rust | Rikare logg | Protokollet får textkategorier (`"kind":"say"`, `"combat"`, `"system"`) – en bakåtkompatibel utökning. Färgad logg, kommandohistorik, scrollning. `HashMap<Kind, Style>`, `VecDeque`. |
| 18 | Go | Strid | Egenskaper, tärningsslag, stridsrundor per tick, död och återuppståndelse. Deterministiska tester med injicerad slump. |
| 19 | Rust | Strid i klienten | HP-mätare, målpanel, blinkande träffar. Tidsstyrd rendering med `event::poll(timeout)` och `Instant`. |
| 20 | Go | Sparande | Spelare och föremål sparas i SQLite (`database/sql` + `modernc.org/sqlite`, ren Go). Lösenord med bcrypt. Migreringar. |
| 21 | Rust | Inställningar och inloggning | Inloggningsskärm, konfigfil (`serde` + `toml`), kommandoradsargument med `clap`, egna feltyper med `thiserror`. |
| 22 | Båda | Tester på riktigt | Fuzzing av `ParseCommand` (`go test -fuzz`), egenskapsbaserade tester i Rust (`proptest`), en bot-klient som spelar mot servern i CI. |

## Fas 4 – Tiles och zoner (planerad)

**Mål:** från rum-graf till tile-värld à la *Caves of Qud*: varje rum blir
en **zon** – en karta på t.ex. 80×25 rutor där du går ruta för ruta. Går du
över kanten hamnar du i nästa zon. Rum-kartan från lektion 10 blir
*världskartan*. Klienten får riktig grafik.

| # | Spår | Lektion | Innehåll |
|---|---|---|---|
| 23 | Go | Zoner | Zoner med tiles och entiteter med `(x, y)`. Protokollet får `zone`-frames (hela zonen) och `delta`-frames (bara ändringar) – du lär dig varför man inte skickar allt varje tick. |
| 24 | Go | Procedurgenerering | Grottor med cellulära automater, gångar med "drunkard's walk", terräng med brus (simplex), allt från en seed så att samma värld kan återskapas. |
| 25 | Rust | Kamera | Visa en del av en stor zon, följ spelaren, scrolla. Koordinatsystem, `saturating_sub`, `clamp`. |
| 26 | Go | Synfält | Rekursiv *shadowcasting* på servern (så att ingen kan fuska) och "fog of war": rutor du sett men inte ser nu ritas mörkare. |
| 27 | Go | Vägsökning | A* för varelser med `container/heap` och generics. Jämför med Dijkstra-kartor (som i *Brogue*). |
| 28 | Rust | Grafisk klient | Ett eget fönster med `macroquad` (enkelt) eller `bevy` (stort ekosystem). Ladda ett CP437-tileset (16×16 glyfer – samma idé som Dwarf Fortress-tilesets) eller ett CC0-paket som Kenneys *1-Bit Pack*. **En trait `Renderer` med två implementationer** – terminal och tiles – och resten av klienten är oförändrad. Här lönar sig beslutet från lektion 10. |
| 29 | Rust | Ljus och liv | Ljuskällor med färgblandning, animerat vatten, partiklar, skärmskakning vid träff. Belysning i *Brogue*-stil. |
| 30 | Rust | Musen | Klicka för att gå (A* på klienten), muspekare som visar "En gammal gran" i en tooltip, högerklicksmeny. |

Alla klientdelar organiseras då som ett **Cargo workspace**: `mud-core`
(protokoll, tillstånd), `mud-tui` och `mud-gfx`. Bra övning i modulär design.

## Fas 5 – En utbyggbar server (planerad)

**Mål:** servern ska aldrig behöva startas om för att få nytt innehåll eller
nya funktioner. Det här är kursens kulmen och knyter ihop allt.

| # | Spår | Lektion | Innehåll |
|---|---|---|---|
| 31 | Go | Omladdning av data | Bevaka världsfilerna (polling eller `fsnotify`), validera (lektion 09!), och byt världen *atomiskt* i spelloopen. Spelare i borttagna rum flyttas till en säker plats. Adminkommandot `@reload`. |
| 32 | Go | Skript i Lua | Inbäddad Lua med `gopher-lua` (ren Go). Kommandon och NPC-beteenden i `.lua`-filer som laddas om live. Sandlåda (inga `os`/`io`), tidsgräns per skript med `context`, och den svåra frågan: *vilket API ska skript få?* Kommandoregistret från lektion 04 blir plötsligt väldigt viktigt. |
| 33 | Rust → Go | Plugins i WebAssembly | Skriv plugins i **Rust**, kompilera till `wasm32-unknown-unknown`, kör dem i Go-servern med `wazero` (ren Go, ingen cgo). Ett eget ABI: pluginet exporterar `on_command(ptr, len)`, servern erbjuder `send(player, text)`. Minnesgränser, "bränsle" mot oändliga loopar, versionering. **Här möts språken på riktigt.** |
| 34 | Go | Händelser och hooks | En händelsebuss: `OnEnterRoom`, `OnSay`, `OnTick`, `OnDeath`. Skript och plugins prenumererar; hooks kan avbryta handlingar ("dörren är låst"). |
| 35 | Go | Copyover | Den klassiska MUD-tekniken från 90-talet: servern startar en *ny version av sig själv* med `exec` och lämnar över de öppna anslutningarna (filbeskrivare via `ExtraFiles`) och spelartillståndet. Spelarna ser bara "Världen skakar till ett ögonblick..." Ny kod – utan att någon kopplas ner. |
| 36 | Go | Bygga i spelet | *Online creation*: `@dig north Mörk grotta`, `@desc ...`, `@exit ...` – ändringar valideras, sparas till världsfilen och syns direkt för alla. |

Tre nivåer av utbyggbarhet, från enklast till kraftfullast:

```
 data (JSON, live-omladdning)  →  skript (Lua)  →  plugins (WASM)  →  ny binär (copyover)
     innehåll                     beteenden        snabb, isolerad     allt annat
                                                   kod i valfritt språk
```

## Fas 6 – Ut i världen (valfria spår)

Välj det som intresserar dig mest:

- **Webbklient:** en WebSocket-gateway i Go och klienten kompilerad från Rust
  till WebAssembly – samma `render`-kod ritar i webbläsarens `<canvas>`.
- **Async Rust:** skriv om klienten med `tokio`. `tokio::select!` är
  förvånansvärt likt Gos `select`. Bygg en lasttestare med 1000 bot-spelare.
- **Riktigt telnet:** IAC-förhandling, NAWS (fönsterstorlek) och **GMCP**, så
  att vanliga MUD-klienter som Mudlet kan visa din karta och statusrad.
- **Drift:** `log/slog`, `pprof`, mätvärden för Prometheus, Docker, TLS,
  skydd mot spam och överbelastning.
- **Rust på servern:** skriv synfälts- eller generatorkoden som ett
  Rust-bibliotek och jämför prestanda och utvecklarupplevelse.

---

## Begreppskarta: Go och Rust sida vid sida

| Idé | Go | Rust | Lektioner |
|---|---|---|---|
| Sammansatt data | `struct` | `struct` | 01, 03 |
| Metoder | `func (s *Server) Serve()` | `impl Server { fn serve(&self) }` | 01, 03 |
| Fel | `error`, `if err != nil` | `Result<T, E>`, `?` | 01, 03, 05, 09, 11 |
| Saknat värde | `nil`, `value, ok :=` | `Option<T>` | 03, 05, 11 |
| Minne | skräpsamlare | ägarskap och lån | 03, 06 |
| Samtidighet | goroutines | OS-trådar (`thread::spawn`), senare `tokio` | 02, 06, (fas 6) |
| Meddelanden | `chan T`, `select` | `mpsc::channel`, `recv`/`try_recv` | 06, 07, 13 |
| Delad data | `sync.Mutex` (frivilligt) | `Mutex<T>` (kompilatorn kräver det) | 05, (14) |
| Abstraktion | gränssnitt (implicita) | traits (explicita) | 06, 09, 10, 11, 13 |
| Generics | `func F[T any]` | `fn f<T: Trait>` | 06, (15, 27) |
| Varianter | konstanter + `switch` | `enum` + `match` | 10, 11 |
| Textformat | `fmt.Stringer` | `Display` | 11 |
| JSON | `encoding/json` + taggar | `serde` + attribut | 09, 10, 11 |
| Tester | `go test`, tabelldrivna | `cargo test`, `tests/` | alla |
| Beroenden | `go get`, `go.mod` | `cargo add`, `Cargo.toml` | 11, 13 |

---

## Läsanvisningar

Kursen förutsätter inga förkunskaper i Go eller Rust, men lite
programmeringsvana (variabler, loopar, funktioner i något språk).
Kompletterande läsning per lektion:

| Lektion | Go: [Tour of Go](https://go.dev/tour) / [Effective Go](https://go.dev/doc/effective_go) | Rust: [The Rust Book](https://doc.rust-lang.org/book/) |
|---|---|---|
| 01 | Basics, Methods, Errors | – |
| 02 | Concurrency: Goroutines | – |
| 03 | – | kap. 1–5 (kom igång, ägarskap, structs) och 9 (felhantering) |
| 04 | More types: slices, maps, function values | – |
| 05 | Effective Go: Errors; paketet `sync` | – |
| 06 | – | kap. 10 (generics, traits) och 16 (trådar, kanaler) |
| 07 | Concurrency: Channels, Select | – |
| 08 | Pointers, Structs | – |
| 09 | `go doc encoding/json`, `go doc embed` | – |
| 10 | Interfaces; `go doc sync/atomic` | – |
| 11 | – | kap. 6 (enums, match), 7 (moduler), [serde.rs](https://serde.rs) |
| 12 | – | kap. 13 (closures, iteratorer) |
| 13 | – | [ratatui.rs](https://ratatui.rs) – tutorials |

Bra referenser när du kör fast: `go doc <paket>`, `rustup doc --std`,
[Go by Example](https://gobyexample.com), [Rust by Example](https://doc.rust-lang.org/rust-by-example/),
och kompilatorns felmeddelanden – särskilt Rusts, som ofta föreslår lösningen.

## Tidsåtgång

Ungefärliga tider för någon som kan programmera men är ny i Go och Rust:

| Del | Tid |
|---|---|
| Fas 1 (01–05) | 8–12 timmar |
| Fas 2 (06–13) | 15–25 timmar |
| Fas 3 | 25–35 timmar |
| Fas 4 | 30–45 timmar |
| Fas 5 | 25–40 timmar |

Rust-lektionerna tar typiskt längre tid än Go-lektionerna. Det är normalt –
lånekontrollen (*the borrow checker*) är en tröskel alla går över.
