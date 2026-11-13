# Kursplan – MUD School

> [🇬🇧 English](../en/CURRICULUM.md) · 🇸🇪 Svenska · [🇯🇵 日本語](../ja/CURRICULUM.md)

> Lär dig **Go** och **Rust** genom att bygga en MUD-server och en matchande
> klient från grunden – från en ekoserver på 20 rader till en utbyggbar
> spelvärld med tile-grafik, skript och plugins som laddas utan omstart – ute
> på internet, spelbar i en webbläsare.

> 🇸🇪 Svenska · [🇬🇧 English](../en/CURRICULUM.md) · [🇯🇵 日本語](../ja/CURRICULUM.md)

Det här dokumentet beskriver *hela* resan: varför den ser ut som den gör,
vad du lär dig i varje steg och vart vi är på väg. Alla 44 lektioner finns
implementerade i `lessons/` (med tester och referenslösningar);
[IDEAS.md](IDEAS.md) har mer att bygga vidare på egen hand.

---

## Innehåll

1. [Pedagogisk idé](#pedagogisk-idé)
2. [Arbetssättet i varje lektion](#arbetssättet-i-varje-lektion)
3. [Översikt](#översikt)
4. [Fas 1 – Grunderna (lektion 01–05)](#fas-1--grunderna-lektion-0105)
5. [Fas 2 – En värld och ett ansikte (lektion 06–13)](#fas-2--en-värld-och-ett-ansikte-lektion-0613)
6. [Fas 3 – En levande värld (lektion 14–22)](#fas-3--en-levande-värld-lektion-1422)
7. [Fas 4 – Tiles och zoner (lektion 23–30)](#fas-4--tiles-och-zoner-lektion-2330)
8. [Fas 5 – En utbyggbar server (lektion 31–36)](#fas-5--en-utbyggbar-server-lektion-3136)
9. [Fas 6 – Ut i världen (lektion 37–44)](#fas-6--ut-i-världen-lektion-3744)
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

### 7. Kursen finns på tre språk

Instruktioner, ledtrådar, kursverktyget, testernas felmeddelanden, spelets
texter och spelvärlden finns på **engelska, svenska och japanska**
(`./mudschool init --lang en|sv|ja`, byt senare med `./mudschool lang`).
Kod och kodkommentarer är på engelska – det är så du kommer att möta kod i
verkligheten, och det gör att alla studenter kan läsa varandras lösningar.
Att all spelartext ligger i en egen fil (`messages.go`) är dessutom ett
första steg mot *internationalisering* av själva spelet.

### 8. Repetition med variation

Kärnidéer återkommer i allt svårare former: kanaler i Go (07) → kanaler i
Rust (06, 13) → actor-mönstret för spelloopen (fas 3) → `tokio::select!`
(44). Felhantering: `error` (01) → egna fel (05) → `errors.Join` (09) →
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
Fas 3  En levande värld      14 ─ 22   spelloop, föremål, NPC:er, strid, sparande
Fas 4  Tiles och zoner       23 ─ 30   zoner à la Caves of Qud, procedurgenerering,
                                       synfält, A*, grafisk klient med tileset
Fas 5  En utbyggbar server   31 ─ 36   omladdning live, Lua-skript, WASM-plugins
                                       i Rust, hooks, copyover, bygga i spelet
Fas 6  Ut i världen          37 ─ 44   introduktion, språk per spelare, WebSocket,
                                       Rust-kärnan i webbläsaren, Docker + Caddy,
                                       skydd, en lasttestare med tokio
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

## Fas 3 – En levande värld (lektion 14–22)

**Mål:** världen lever även när ingen skriver något. Föremål, varelser,
strid – och spelare som finns kvar efter en omstart.

| # | Spår | Lektion | Spelet får | Språket lär dig |
|---|---|---|---|---|
| 14 | Go | Spelloopen | en värld som tickar, en snygg nedstängning med Ctrl+C | actor-mönstret (en goroutine äger världen), `time.Ticker`, `select`, `context.Context`, `signal.NotifyContext` |
| 15 | Go | Föremål | `get`, `drop`, `inventory`, `examine`, föremål i världsfilen | generics (`FindMatch[T Matcher]`, `remove[T comparable]`), gränssnitt som villkor, `go:embed` |
| 16 | Go | Varelser | NPC:er som vandrar och pratar | ett `Behavior`-gränssnitt, en JSON-"tagged union", `math/rand/v2` med fast seed |
| 17 | Rust | Rikare logg | textsorter i färg, kommandohistorik, scrollning | enums med `#[serde(other)]` och `#[default]`, `Copy`, `std::mem::take` |
| 18 | Go | Strid | `kill`, `flee`, rundor per tick, död, varelser som kommer tillbaka | tärningar med injicerad slump, tester som stegar klockan själva |
| 19 | Rust | Strid i klienten | en målpanel med hälsomätare, en statusrad som blinkar rött | `Instant`/`Duration`, en injicerad klocka, `Gauge`, `recv_timeout` |
| 20 | Go | Spara | konton med lösenord; rum, hälsa och föremål överlever en omstart | `database/sql`, SQLite i ren Go, transaktioner, migreringar, bcrypt, moduler (`go get`) |
| 21 | Rust | Inställningar och inloggning | `--name`, `--help`, `mudclient.toml`, ett dolt lösenordsfält | `clap` derive, `toml` + serde, `thiserror`, `ExitCode` |
| 22 | Go | Testa på riktigt | `ParseCommand` som klarar alla sorters blanktecken | fuzzing (`go test -fuzz`), invarianter, egenskapsbaserade tester, `go/parser` |

Proptest i Rust och en bot-klient är utmaningar i lektion 22.

**Efter fas 3** lever världen: saker händer medan du sitter still, och den
minns dig nästa dag.

## Fas 4 – Tiles och zoner (lektion 23–30)

**Mål:** från rum-graf till tile-värld à la *Caves of Qud*: varje rum blir
en **zon** – en karta på t.ex. 31×13 rutor där du går ruta för ruta. Går du
över kanten hamnar du i nästa zon. Rum-kartan från lektion 10 blir
*världskartan*. Klienten får riktig grafik.

| # | Spår | Lektion | Spelet får | Språket lär dig |
|---|---|---|---|---|
| 23 | Go | Zoner | varje rum ett rutnät; steg för steg; handgjorda zoner i världsfilen; `zone`- och `delta`-frames | 2D-rutnät, jämförbara structs som nycklar, skillnader i tillstånd |
| 24 | Go | Procedurgenerering | grottor (cellautomater), skogar och sjöar (värdebrus), gångar (den fulle mannens vandring), allt från ett frö | `math/rand/v2` med frö, `hash/fnv`, flood fill, egenskapstester |
| 25 | Rust | Kameran | zonen på skärmen, en kamera som följer dig, gå med piltangenterna | serde med `rename`, `From`, `BTreeMap`, `saturating_sub`, match-villkor |
| 26 | Go | Synfält | rekursiv skuggkastning på servern, krigets dimma, bara synliga entiteter skickas | rekursion, symmetri, skilja på vad servern vet och vad en klient får veta |
| 27 | Go | Vägval | vargar som jagar dig, råttor som flyr | generisk A* (`[N comparable]`), `container/heap`, Dijkstra-kartor |
| 28 | Rust | En grafisk klient | ett fönster med kursens eget tileset; krigets dimma i båda klienterna | en trait `Renderer` med flera implementationer, Cargo-features, ett andra program, macroquad |
| 29 | Rust | Ljus och liv | facklor, mörka grottor, glittrande vatten, gnistor och skakningar | ljus som tal, animation som rena funktioner av tiden, partiklar, standardmetoder |
| 30 | Rust | Musen | klicka för att gå, verktygstips, en högerklicksmeny | A* med `BinaryHeap` + `Reverse`, koordinatsystem, en kö av handlingar över tid |

Att organisera klienten som ett **Cargo workspace** (`mud-core`, `mud-tui`,
`mud-gfx`) är en utmaning i lektion 28.

**Efter fas 4** ser MUD:en ut som en roguelike: genererade grottor och
skogar, saker du bara ser när du ser dem, och ett fönster med tiles.

## Fas 5 – En utbyggbar server (lektion 31–36)

**Mål:** servern ska aldrig behöva startas om för att få nytt innehåll eller
nya funktioner. Den knyter ihop allt från fas 1–4.

| # | Spår | Lektion | Spelet får | Språket lär dig |
|---|---|---|---|---|
| 31 | Go | Ladda om världen | världsfilen bevakas och byts i drift; `@reload`; administratörer | pollning med `os.Stat` och en ticker, validera utanför loopen, byt inne i den, flytta levande tillstånd |
| 32 | Go | Skript i Lua | kommandon och varelsebeteenden i `.lua`-filer, omladdade när du sparar | bädda in gopher-lua, en sandlåda, tidsgränser med `context`, API-design |
| 33 | Go (+ Rust) | Plugins i WebAssembly | ett plugin skrivet i Rust, kompilerat till `wasm32`, som körs i Go-servern | wazero, ett eget ABI, delat minne, minnes- och tidsgränser, versioner |
| 34 | Go | Händelser och hooks | `enter`, `say`, `tick`, `death`; hooks i Go och Lua som kan säga nej | en händelsebuss, funktioner som prenumeranter, returvärden från Lua |
| 35 | Go | Copyover | ett nytt program tar över utan att någon kopplas ner | fildeskriptorer, `ExtraFiles`, `net.FileConn`/`FileListener`, överlämning av tillstånd |
| 36 | Go | Bygga i spelet | `@dig`, `@desc`, `@name`, `@exit`, validerat och sparat direkt | genomför–validera–ångra, atomiska skrivningar, ordning på samtidiga sparningar |

Tre nivåer av utbyggbarhet, från enklast till kraftfullast:

```
 data (JSON, live-omladdning)  →  skript (Lua)  →  plugins (WASM)  →  ny binär (copyover)
     innehåll                     beteenden        snabb, isolerad     allt annat
                                                   kod i valfritt språk
```

## Fas 6 – Ut i världen (lektion 37–44)

**Mål:** MUD:en lämnar din dator. Nya spelare tas emot, alla läser världen på
sitt eget språk, vem som helst kan spela i en webbläsare, och servern kör på
en VPS – skyddad, mätt och lasttestad.

| # | Spår | Lektion | Spelet får | Språket lär dig |
|---|---|---|---|---|
| 37 | Go | Introduktionen | nya spelare börjar i ett avskilt område och leds steg för steg; hur långt de kommit sparas | beteende som data, en hook för `EventCommand`, en databasmigrering, bredden-först-sökning |
| 38 | Go | Språk per spelare | `language`; introduktionen och översatta rum på varje spelares språk | `json.Unmarshaler`/`Marshaler`, metoder på en map-typ, `text/template`, reservspråk |
| 39 | Go | WebSocket | en webbklient som servern serverar; webbspelare möter telnetspelare | `net/http`, `http.Handler`, en `net.Listener` av kanaler, `go:embed`, Origin-kontrollen |
| 40 | Rust | Kärnan | klienten blir en Cargo-workspace med `mud-core`: ingen I/O, ingen klocka | workspaces, `pub use`, orphan rule, att skicka in tiden |
| 41 | Rust | Klienten i webbläsaren | kärnan, kompilerad till WebAssembly, ritar zonen på en `<canvas>`; klicka för att gå | `cdylib`, `#[no_mangle]`, `extern "C"`, `thread_local!`, ett ABI med webbläsaren som värd |
| 42 | Go | Ut på nätet | Docker, Caddy med automatisk HTTPS, en guide för en VPS | konfiguration från miljön och en testbar `main`, `log/slog`, `/healthz`, SIGTERM, `X-Forwarded-For` |
| 43 | Go | Skydd | gränser för kommandon, anslutningar per IP, långa rader och långsamma inloggningar; räknare | en token bucket, läs-tidsgränser, `expvar`, `pprof` |
| 44 | Rust | Lasttestaren | hundratals botar som går omkring och mäter svarstider | tokio, `async`/`.await`, `tokio::select!`, `mpsc`/`watch`, percentiler |

```
webbläsare ──https──▶ Caddy ──▶ gateway (/ws) ──▶ ConnListener ──▶ Serve ◀── telnet
   └── mud_web.wasm: Rust-kärnan ritar zonen, samma kod som fönstret
```

Fler idéer – GMCP för Mudlet, Rust på servern, en asynkron terminalklient –
finns i [IDEAS.md](IDEAS.md).

---

## Begreppskarta: Go och Rust sida vid sida

| Idé | Go | Rust | Lektioner |
|---|---|---|---|
| Sammansatt data | `struct` | `struct` | 01, 03 |
| Metoder | `func (s *Server) Serve()` | `impl Server { fn serve(&self) }` | 01, 03 |
| Fel | `error`, `if err != nil` | `Result<T, E>`, `?` | 01, 03, 05, 09, 11 |
| Saknat värde | `nil`, `value, ok :=` | `Option<T>` | 03, 05, 11 |
| Minne | skräpsamlare | ägarskap och lån | 03, 06 |
| Samtidighet | goroutines | OS-trådar (`thread::spawn`), senare `tokio`-tasks | 02, 06, 44 |
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
programmeringsvana (variabler, loopar, funktioner i något språk). Hur du
installerar allt står i [SETUP.md](SETUP.md).
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
| Fas 3 (14–22) | 25–35 timmar |
| Fas 4 (23–30) | 30–45 timmar |
| Fas 5 (31–36) | 25–40 timmar |
| Fas 6 (37–44) | 30–45 timmar |

Rust-lektionerna tar typiskt längre tid än Go-lektionerna. Det är normalt –
lånekontrollen (*the borrow checker*) är en tröskel alla går över.
