# MUD School 🐉

> 🇸🇪 Svenska · [🇬🇧 English](README.en.md) · [🇯🇵 日本語](README.ja.md)

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
då blir kartan ren ASCII.)

Sedan fortsätter kursplanen: strid och NPC:er, zoner med tiles och
procedurgenerering à la Caves of Qud, en grafisk klient med tileset, och en
server som kan byggas ut **medan den kör** – med omladdning av världen,
Lua-skript, plugins skrivna i Rust och kompilerade till WebAssembly, och
den klassiska MUD-tekniken *copyover*.

Kursen finns på **svenska, engelska och japanska** – instruktioner, ledtrådar,
testernas felmeddelanden, spelets texter och spelvärlden.

## Kom igång

**1. Installera verktygen.** Du behöver Go 1.22+, Rust (stable) och en
terminal med UTF-8 och färg – telnet eller nc är bra att ha. Steg för steg
för macOS, Linux och Windows finns i **[docs/sv/SETUP.md](docs/sv/SETUP.md)**.

Snabbast: öppna repot i **GitHub Codespaces** eller en **VS Code Dev
Container** – allt är förinstallerat (`.devcontainer/`).

**2. Kontrollera och börja.**

```
git clone https://github.com/mattias/mud-school.git && cd mud-school
./mudschool doctor             # kontrollerar att allt är installerat
./mudschool init --lang sv     # skapar workspace/ med din egen kod (sv | en | ja)
./mudschool show               # läs första lektionen
./mudschool watch              # testerna körs varje gång du sparar
```

På Windows utan WSL: se [SETUP.md](docs/sv/SETUP.md#c-windows).

## Kursverktyget

| Kommando | Vad det gör |
|---|---|
| `./mudschool doctor` | kontrollerar Go, Rust, telnet/nc, UTF-8 och färg |
| `./mudschool list` | alla lektioner och din status |
| `./mudschool show [id]` | instruktionerna för en lektion |
| `./mudschool check [--race]` | kör lektionens tester; vid grönt går du vidare |
| `./mudschool watch` | som `check`, men automatiskt när du sparar |
| `./mudschool hint` | nästa ledtråd (en i taget) |
| `./mudschool run server` / `run client` | starta din server eller klient |
| `./mudschool solution [--apply]` | visa (eller kopiera in) referenslösningen |
| `./mudschool goto <id>` | hoppa till en lektion |
| `./mudschool lang sv\|en\|ja` | byt språk |

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

Fas 3–6 (lektion 14–36 och valfria spår) beskrivs i **[docs/sv/CURRICULUM.md](docs/sv/CURRICULUM.md)**.

## Dokumentation

| | 🇸🇪 | 🇬🇧 | 🇯🇵 |
|---|---|---|---|
| Installation och miljö | [SETUP](docs/sv/SETUP.md) | [SETUP](docs/en/SETUP.md) | [SETUP](docs/ja/SETUP.md) |
| Kursplan: pedagogik, alla faser, Go ↔ Rust | [CURRICULUM](docs/sv/CURRICULUM.md) | [CURRICULUM](docs/en/CURRICULUM.md) | [CURRICULUM](docs/ja/CURRICULUM.md) |
| Protokollet mellan server och klient | [PROTOCOL](docs/sv/PROTOCOL.md) | [PROTOCOL](docs/en/PROTOCOL.md) | [PROTOCOL](docs/ja/PROTOCOL.md) |
| Idéer för spelet, klienten och kursen | [IDEAS](docs/sv/IDEAS.md) | [IDEAS](docs/en/IDEAS.md) | [IDEAS](docs/ja/IDEAS.md) |
| För lärare: hur kursen är byggd | [TEACHERS](docs/sv/TEACHERS.md) | [TEACHERS](docs/en/TEACHERS.md) | [TEACHERS](docs/ja/TEACHERS.md) |

## För lärare

`./mudschool verify` bygger upp referenslösningen lektion för lektion, på alla
tre språken, och kontrollerar att varje lektions tester **misslyckas före** och
**går igenom efter** lösningen. `./mudschool reference <katalog> [id]` ger dig
hela referenslösningen till och med en viss lektion. CI kör `verify` vid varje push.
