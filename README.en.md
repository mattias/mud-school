# MUD School 🐉

> [🇸🇪 Svenska](README.md) · 🇬🇧 English · [🇯🇵 日本語](README.ja.md)

**Learn Go and Rust by building your own online game – from scratch.**

You build a **MUD server in Go** and a **MUD client in Rust**, step by step.
It all starts with a 20-line echo server. After lesson 13 you have a
multiplayer world with 21 rooms, chat and a JSON protocol – and a terminal
client with a Dwarf Fortress-style colour map, a status bar and a log:

```
 Anna · The Crossroads · ♥♥♥♥♥♥♥♥♥♥ HP 10/10 · 2 online
┌ Log ──────────────────────────────────────────────────────────┐┌ Map ────────┐
│> s                                                            ││      #      │
│The Town Gate                                                  ││      │      │
│A heavy wooden gate in the town wall. To the south a road winds││      ·      │
│Exits: north, south                                            ││      │      │
│> s                                                            ││      ·   "  │
│The Country Road                                               ││      │   │  │
│A muddy road with deep wheel ruts. Mudby's walls are visible to││≈─"─"─@─♣─♣  │
│Exits: north, south                                            ││      │   │  │
│> s                                                            ││      ·   ♣  │
│The Country Road                                               ││      │      │
│The road is lined with low stone walls and wild roses.         ││      ▲─Ω    │
│Exits: north, south                                            ││        │    │
│> s                                                            ││        >    │
│The Crossroads                                                 ││             │
│A signpost points in every direction. The writing is too weathe││             │
│Exits: north, east, south, west                                ││             │
└───────────────────────────────────────────────────────────────┘└─────────────┘
┌ Command ─────────────────────────────────────────────────────────────────────┐
│>                                                                             │
└──────────────────────────────────────────────────────────────────────────────┘
```

(This is a real screenshot of the reference solution. Telnet works too; then
the map is plain ASCII.)

Then the curriculum continues: combat and NPCs, zones with tiles and
procedural generation à la Caves of Qud, a graphical client with a tileset,
and a server that can be extended **while it is running** – with live world
reloading, Lua scripts, plugins written in Rust and compiled to WebAssembly,
and the classic MUD technique *copyover*.

The course is available in **Swedish, English and Japanese** – instructions,
hints, the tests' failure messages, the game's texts and the game world.

## Getting started

**1. Install the tools.** You need Go 1.22+, Rust (stable) and a terminal with
UTF-8 and colour – telnet or nc are good to have. Step by step for macOS,
Linux and Windows: **[docs/en/SETUP.md](docs/en/SETUP.md)**.

Fastest: open the repository in **GitHub Codespaces** or a **VS Code Dev
Container** – everything is preinstalled (`.devcontainer/`).

**2. Check and start.**

```
git clone https://github.com/mattias/mud-school.git && cd mud-school
./mudschool doctor             # checks that everything is installed
./mudschool init --lang en     # creates workspace/ with your own code (sv | en | ja)
./mudschool show               # read the first lesson
./mudschool watch              # the tests run every time you save
```

On Windows without WSL: see [SETUP.md](docs/en/SETUP.md#c-windows).

## The course tool

| Command | What it does |
|---|---|
| `./mudschool doctor` | checks Go, Rust, telnet/nc, UTF-8 and colour |
| `./mudschool list` | every lesson and your progress |
| `./mudschool show [id]` | the instructions for a lesson |
| `./mudschool check [--race]` | runs the lesson's tests; when green you move on |
| `./mudschool watch` | like `check`, but automatically when you save |
| `./mudschool hint` | the next hint (one at a time) |
| `./mudschool run server` / `run client` | start your server or client |
| `./mudschool solution [--apply]` | show (or copy in) the reference solution |
| `./mudschool goto <id>` | jump to a lesson |
| `./mudschool lang sv\|en\|ja` | switch language |

All your code lives in `workspace/` – `server/` (Go) and `client/` (Rust).
You can also run `go test ./...` and `cargo test` directly there.

## The lessons

| # | | Lesson | You learn |
|---|---|---|---|
| 01 | Go | Hello, TCP! | packages, structs, methods, `error`, `defer`, `net` |
| 02 | Go | Many players at once | goroutines, the race detector |
| 03 | Rust | Hello, Rust! – a first client | Cargo, `Result`/`?`, `Option`, ownership and borrowing |
| 04 | Go | Commands | maps, slices, functions as values, table-driven tests |
| 05 | Go | Players | your own errors, UTF-8 and runes, `sync.Mutex` |
| 06 | Rust | Threads and channels | `thread`, `mpsc`, generics, `Send`/`'static` |
| 07 | Go | Chat | channels, `select`, protection against slow clients |
| 08 | Go | Rooms | pointers, closures, a world as a graph |
| 09 | Go | World files | JSON, validation, `errors.Join`, `go:embed` |
| 10 | Go | Protocol and map | interfaces, two encodings of the same game, `sync/atomic` |
| 11 | Rust | The protocol in Rust | enums with data, `match`, serde, `Display` |
| 12 | Rust | The map in colour | iterators, `Copy`, ANSI truecolor |
| 13 | Rust | A real user interface | ratatui, an event loop, testing a UI |

Phases 3–6 (lessons 14–36 and optional tracks) are described in **[docs/en/CURRICULUM.md](docs/en/CURRICULUM.md)**.

## Documentation

| | 🇸🇪 | 🇬🇧 | 🇯🇵 |
|---|---|---|---|
| Setup and environment | [SETUP](docs/sv/SETUP.md) | [SETUP](docs/en/SETUP.md) | [SETUP](docs/ja/SETUP.md) |
| Curriculum: teaching, every phase, Go ↔ Rust | [CURRICULUM](docs/sv/CURRICULUM.md) | [CURRICULUM](docs/en/CURRICULUM.md) | [CURRICULUM](docs/ja/CURRICULUM.md) |
| The protocol between server and client | [PROTOCOL](docs/sv/PROTOCOL.md) | [PROTOCOL](docs/en/PROTOCOL.md) | [PROTOCOL](docs/ja/PROTOCOL.md) |
| Ideas for the game, the client and the course | [IDEAS](docs/sv/IDEAS.md) | [IDEAS](docs/en/IDEAS.md) | [IDEAS](docs/ja/IDEAS.md) |
| For teachers: how the course is built | [TEACHERS](docs/sv/TEACHERS.md) | [TEACHERS](docs/en/TEACHERS.md) | [TEACHERS](docs/ja/TEACHERS.md) |

## For teachers

`./mudschool verify` builds up the reference solution lesson by lesson, in all
three languages, and checks that every lesson's tests **fail before** and
**pass after** the solution. `./mudschool reference <dir> [id]` gives you the
whole reference solution up to a given lesson. CI runs `verify` on every push.
