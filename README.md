# MUD School 🐉

**Lär dig Go och Rust genom att bygga ett eget onlinespel – från grunden.**

Du bygger en **MUD-server i Go** och en **MUD-klient i Rust**, steg för steg.
Allt börjar med en ekoserver på 20 rader. Efter lektion 13 har du en
flerspelarvärld med 21 rum, chatt och ett JSON-protokoll – och en
terminalklient med färgkarta i Dwarf Fortress-stil, statusrad och logg:

```
 Anna · Vägkorsningen · ♥♥♥♥♥♥♥♥♥♥ HP 10/10 · 2 online
┌ Logg ───────────────────────────────────────────────────────┐┌ Karta ──────┐
│Bertil har anslutit.                                         ││      #      │
│Bertil säger: Hej Anna!                                      ││      │      │
│> s                                                          ││      ·      │
│Landsvägen                                                   ││      │      │
│Vägen kantas av låga stenmurar och vildrosor.                ││      ·   "  │
│Utgångar: norr, söder                                        ││      │   │  │
│> s                                                          ││≈─"─"─@─♣─♣  │
│Vägkorsningen                                                ││      │   │  │
│En vägvisare pekar åt alla håll. Texten är för vittrad för   ││      ·   ♣  │
│att läsa.                                                    ││      │      │
│Utgångar: norr, öster, söder, väster                         ││      ▲─Ω    │
└─────────────────────────────────────────────────────────────┘└─────────────┘
┌ Kommando ──────────────────────────────────────────────────────────────────┐
│> say hej Bertil_                                                           │
└────────────────────────────────────────────────────────────────────────────┘
```

(Det här är en riktig skärmbild av referenslösningen. Telnet fungerar också,
så då blir kartan ren ASCII.)

Sedan fortsätter kursplanen: strid och NPC:er, zoner med tiles och
procedurgenerering à la Caves of Qud, en grafisk klient med tileset, och en
server som kan byggas ut **medan den kör** – med omladdning av världen,
Lua-skript, plugins skrivna i Rust och kompilerade till WebAssembly, och
den klassiska MUD-tekniken *copyover*.

## Kom igång

Du behöver:

- [Go](https://go.dev/dl/) 1.22 eller senare
- [Rust](https://rustup.rs) (stabil, 2021-utgåvan)
- `telnet` eller `nc` (netcat) för att prova servern
- en terminal med UTF-8 och truecolor (de flesta moderna har det)
- en editor med `gopls` och `rust-analyzer` (t.ex. VS Code) rekommenderas

```
git clone <det här repot> && cd mud-school
./mudschool init      # skapar workspace/ med din egen kod
./mudschool show      # läs första lektionen
./mudschool watch     # testerna körs varje gång du sparar
```

På Windows: kör kommandona i WSL eller Git Bash, eller direkt med
`go run ./tools/mudschool <kommando>` från `tools/mudschool`.

## Kursverktyget

| Kommando | Vad det gör |
|---|---|
| `./mudschool list` | alla lektioner och din status |
| `./mudschool show [id]` | instruktionerna för en lektion |
| `./mudschool check [--race]` | kör lektionens tester; vid grönt går du vidare |
| `./mudschool watch` | som `check`, men automatiskt när du sparar |
| `./mudschool hint` | nästa ledtråd (en i taget) |
| `./mudschool run server` / `run client` | starta din server eller klient |
| `./mudschool solution [--apply]` | visa (eller kopiera in) referenslösningen |
| `./mudschool goto <id>` | hoppa till en lektion |

All din kod ligger i `workspace/` – `server/` (Go) och `client/` (Rust).
Du kan också köra `go test ./...` och `cargo test` direkt där.

## Lektionerna

| # | | Lektion | Du lär dig |
|---|---|---|---|
| 01 | Go | Hej, TCP! | paket, structs, metoder, `error`, `defer`, `net` |
| 02 | Go | Många spelare samtidigt | goroutines, race detectorn |
| 03 | Rust | Hej, Rust! – en första klient | Cargo, `Result`/`?`, `Option`, ägarskap och lån |
| 04 | Go | Kommandon | maps, slices, funktioner som värden, tabelldrivna tester |
| 05 | Go | Spelare | egna fel, UTF-8 och runor, `sync.Mutex` |
| 06 | Rust | Trådar och kanaler | `thread`, `mpsc`, generics, `Send`/`'static` |
| 07 | Go | Chatt | kanaler, `select`, skydd mot långsamma klienter |
| 08 | Go | Rum | pekare, closures, en värld som graf |
| 09 | Go | Världsfiler | JSON, validering, `errors.Join`, `go:embed` |
| 10 | Go | Protokoll och karta | gränssnitt, två kodningar av samma spel, `sync/atomic` |
| 11 | Rust | Protokoll i Rust | enums med data, `match`, serde, `Display` |
| 12 | Rust | Kartan i färg | iteratorer, `Copy`, ANSI truecolor |
| 13 | Rust | Ett riktigt gränssnitt | ratatui, händelseloop, testa UI |

Fas 3–6 (lektion 14–36 och valfria spår) beskrivs i **[docs/KURSPLAN.md](docs/KURSPLAN.md)**.

## Dokumentation

- **[docs/KURSPLAN.md](docs/KURSPLAN.md)** – den pedagogiska idén, alla faser, begreppskarta Go ↔ Rust, läsanvisningar
- **[docs/PROTOKOLL.md](docs/PROTOKOLL.md)** – protokollet mellan server och klient
- **[docs/IDEER.md](docs/IDEER.md)** – idéer för spelet, klienten och kursen
- **[docs/FOR-LARARE.md](docs/FOR-LARARE.md)** – hur kursen är byggd och hur du lägger till lektioner

## För lärare

`./mudschool verify` bygger upp referenslösningen lektion för lektion och
kontrollerar att varje lektions tester **misslyckas före** och **går igenom
efter** lösningen. `./mudschool reference <katalog> [id]` ger dig hela
referenslösningen till och med en viss lektion. CI kör `verify` vid varje push.
