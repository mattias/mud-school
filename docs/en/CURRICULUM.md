# Curriculum – MUD School

> 🇬🇧 English · [🇸🇪 Svenska](../sv/CURRICULUM.md) · [🇯🇵 日本語](../ja/CURRICULUM.md)

> Learn **Go** and **Rust** by building a MUD server and a matching client
> from scratch – from a 20-line echo server to an extensible game world with
> tile graphics, scripts and plugins that load without a restart.

This document describes the *whole* journey: why it looks the way it does,
what you learn in every step and where we are heading. Lessons 01–22 are
implemented in `lessons/` (with tests and reference solutions). Phases 4–6 are
planned and described here so that you can see the big picture – and build on
your own.

---

## Contents

1. [The teaching idea](#the-teaching-idea)
2. [How every lesson works](#how-every-lesson-works)
3. [Overview](#overview)
4. [Phase 1 – The basics (lessons 01–05)](#phase-1--the-basics-lessons-0105)
5. [Phase 2 – A world and a face (lessons 06–13)](#phase-2--a-world-and-a-face-lessons-0613)
6. [Phase 3 – A living world (lessons 14–22)](#phase-3--a-living-world-lessons-1422)
7. [Phase 4 – Tiles and zones](#phase-4--tiles-and-zones-planned)
8. [Phase 5 – An extensible server](#phase-5--an-extensible-server-planned)
9. [Phase 6 – Out into the world](#phase-6--out-into-the-world-optional-tracks)
10. [Concept map: Go and Rust side by side](#concept-map-go-and-rust-side-by-side)
11. [Further reading](#further-reading)
12. [Time required](#time-required)

---

## The teaching idea

### 1. A game that grows – not exercises that disappear

You build **one** program throughout the course. The code from lesson 04 is
still alive in lesson 13. That means you have to live with your own
decisions and refactor when the world changes – just like in real projects.
Every lesson ends with something you can *run and show* to someone.

### 2. Concepts exactly when the problem needs them (just-in-time)

No language concept is introduced "because it is in the book". Every concept
shows up when the game needs it:

| The problem | The concept |
|---|---|
| The second player can't get in | goroutines (02) |
| Two players log in at once and the server crashes | `sync.Mutex`, data races (05) |
| A player with a bad connection freezes the whole server | channels, `select`, one writer per socket (07) |
| The client misses messages while you type | threads and channels in Rust (06) |
| The map should be green in one client and an image in another | interfaces, separating meaning from appearance (10) |
| The server sends different kinds of messages | enums with data, `match` (11) |
| 600 rooms can't be written in code | JSON, validation, `go:embed` (09) |

That way you understand *why* a concept exists – not just how it is written.

### 3. Two languages, on purpose

**Go first.** Go is small and forgiving, and network programming is its home
turf. You have a working multiplayer server after two lessons. That early
success carries you through the harder parts.

**Rust when there is something to talk to.** Rust is introduced in lesson 03,
when the server already exists. Rust is harder, but you learn it with a
concrete task and something to test against (your own server).

**Contrastive learning.** Every Rust lesson has a *Go ↔ Rust* comparison.
Seeing the same idea (a channel, an error, an interface) expressed in two ways
makes you understand the idea better than if you only knew one language. You
learn to separate *what* (concurrency, error handling) from *how* (goroutines
vs threads, `error` vs `Result`).

**The right tool for the job.** The split is not arbitrary: Go for a network
server with thousands of concurrent connections, Rust for a client with
graphics and full control. In phase 5 the languages meet: plugins written in
Rust, compiled to WebAssembly, run inside the Go server.

### 4. The tests are the specification

Every lesson has tests that `./mudschool check` runs. They are written to be
readable: the failure messages say what was expected and what happened, in
your language. The workflow becomes red → green → refactor.
`./mudschool watch` runs the tests every time you save, so feedback arrives
within seconds.

Many of the server's tests are **black-box tests over TCP**: they connect to
your server exactly like a real client. That means you are free to structure
your code however you like – as long as it behaves correctly.

### 5. Support that fades (scaffolding → fading)

| Phase | You get | You write |
|---|---|---|
| 1 | skeletons with `TODO` and almost-finished function signatures | the function bodies |
| 2 | new files with types and signatures, a description of the rest | functions, changes to existing code |
| 3–4 | a contract (protocol, interfaces) and tests | the design, and your own tests beyond the course's |
| 5–6 | a goal and some guidelines | everything, including the tests |

### 6. Help in the right doses

- `./mudschool hint` gives hints **one at a time** – from a nudge in the
  right direction to almost-code. Try yourself first.
- `./mudschool solution` shows the reference solution. Use it to *compare
  afterwards* (`diff -ru`), not to copy.
- **Think about** – reflection questions in every lesson. Answer them
  (preferably in writing, or out loud to a study partner). That is where the
  understanding settles.
- **Challenges** – optional extensions for those who want more. They don't
  affect the tests.

### 7. The course is available in three languages

Instructions, hints, the course tool, the tests' failure messages, the game's
texts and the game world are available in **English, Swedish and Japanese**
(`./mudschool init --lang en|sv|ja`, change later with `./mudschool lang`).
Code and code comments are in English – that is how you will meet code in
the real world, and it lets every student read each other's solutions. Having
all player-facing text in a file of its own (`messages.go`) is also a first
step towards *internationalizing* the game itself.

### 8. Repetition with variation

Core ideas come back in increasingly demanding forms: channels in Go (07) →
channels in Rust (06, 13) → the actor pattern for the game loop (phase 3) →
`tokio::select!` (phase 6). Error handling: `error` (01) → your own errors
(05) → `errors.Join` (09) → `Result` and `?` (03) → your own error types (11).
Every time you see an idea again you understand it a little more deeply.

---

## How every lesson works

```
./mudschool show        # 1. read: goals, concepts, task
                        # 2. write code in workspace/
./mudschool watch       # 3. the tests run every time you save
./mudschool hint        # 4. stuck? one hint at a time
./mudschool run server  # 5. try it for real (telnet, your client)
                        # 6. answer "Think about", try a challenge
                        # 7. compare with the reference solution
```

Every README follows the same template: **Goals → concepts with small
examples → Task → Try it yourself → Think about → Challenges**.

---

## Overview

```
Phase 1  The basics           01 ─ 05   Go: TCP, goroutines, commands, players
                                        Rust: the first client
Phase 2  A world and a face   06 ─ 13   Go: chat, rooms, world files, protocol + map
                                        Rust: threads, serde, colour map, TUI with ratatui
Phase 3  A living world       14 ─ 22   game loop, items, NPCs, combat, saving
───────────────────────────────────────────────────── implemented above ───
Phase 4  Tiles and zones      23 ─ 30   Caves of Qud-style zones, procedural generation,
                                        field of view, A*, graphical client with a tileset
Phase 5  An extensible server 31 ─ 36   live reloading, Lua scripts, WASM plugins
                                        in Rust, hooks, copyover, building in the game
Phase 6  Out into the world   optional  web client (Rust→WASM), async Rust,
                                        GMCP, operations and load testing
```

---

## Phase 1 – The basics (lessons 01–05)

**Goal:** a multiplayer server you can log in to with telnet, and a first
client in Rust. You can read and write simple programs in both languages.

| # | Track | Lesson | The game gets | The language teaches you |
|---|---|---|---|---|
| 01 | Go | Hello, TCP! | an echo server | packages, functions, methods, structs, `error`, `defer`, `net`, `bufio` |
| 02 | Go | Many players at once | several concurrent players | goroutines, closures, `sync.WaitGroup`, the race detector |
| 03 | Rust | Hello, Rust! | a client of its own | Cargo, modules, `struct`/`impl`, `Result`/`?`, `Option`, ownership and borrowing, `String`/`&str` |
| 04 | Go | Commands | `look`, `help`, abbreviations, a command registry | `strings`, maps, slices, functions as values, table-driven tests |
| 05 | Go | Players | login, unique names, `who` | your own errors, `errors.Is`, `%w`, runes and UTF-8, `sync.Mutex` |

## Phase 2 – A world and a face (lessons 06–13)

**Goal:** a world to explore, a protocol for graphical clients and a client
with a colour map, status bar and log in the terminal.

| # | Track | Lesson | The game gets | The language teaches you |
|---|---|---|---|---|
| 06 | Rust | Threads and channels | the client listens and types at the same time | `thread::spawn`, `move`, `mpsc`, generics, trait bounds, `Send`/`'static` |
| 07 | Go | Chat | `say`, logins/logouts announced, protection against slow clients | channels, buffered channels, `select`/`default`, backpressure |
| 08 | Go | Rooms | rooms, exits, `look`, movement, `shout` | pointers, variadic functions, closures as factories, lock conventions |
| 09 | Go | World files | 21 rooms in JSON, validation, a `-world` flag | `encoding/json`, struct tags, `io.Reader`, `errors.Join`, `go:embed` |
| 10 | Go | Protocol and map | a JSON protocol, a map, a status – and telnet still works | interfaces/encodings, `sync/atomic`, 2D grids |
| 11 | Rust | The protocol in Rust | the client understands JSON | enums with data, `match`, `let-else`, serde, `Display` |
| 12 | Rust | The map in colour | coloured tiles in Dwarf Fortress style | `Copy`, iterators, `fmt::Write`, ANSI codes |
| 13 | Rust | A real user interface | TUI: status bar, log, map, input | ratatui, `From`/`Into`, an event enum, `TestBackend` |

**After phase 2** you have a real little MUD. It is a good place to pause:
show someone, rewrite a part you are not happy with, do some challenges.

---

## Phase 3 – A living world (lessons 14–22)

**Goal:** the world lives even when nobody types anything. Items, creatures,
combat – and players who remain after a restart.

| # | Track | Lesson | The game gets | The language teaches you |
|---|---|---|---|---|
| 14 | Go | The game loop | a world that ticks, a clean shutdown with Ctrl+C | the actor pattern (one goroutine owns the world), `time.Ticker`, `select`, `context.Context`, `signal.NotifyContext` |
| 15 | Go | Items | `get`, `drop`, `inventory`, `examine`, items in the world file | generics (`FindMatch[T Matcher]`, `remove[T comparable]`), interfaces as constraints, `go:embed` |
| 16 | Go | Creatures | NPCs that wander and talk | a `Behavior` interface, a JSON "tagged union", `math/rand/v2` with a fixed seed |
| 17 | Rust | A richer log | text kinds in colour, command history, scrolling | enums with `#[serde(other)]` and `#[default]`, `Copy`, `std::mem::take` |
| 18 | Go | Combat | `kill`, `flee`, rounds per tick, dying, creatures that come back | dice with injected randomness, tests that step the clock themselves |
| 19 | Rust | Combat in the client | a target panel with a health bar, a status bar that flashes red | `Instant`/`Duration`, an injected clock, `Gauge`, `recv_timeout` |
| 20 | Go | Saving | accounts with passwords; rooms, health and items survive a restart | `database/sql`, SQLite in pure Go, transactions, migrations, bcrypt, modules (`go get`) |
| 21 | Rust | Settings and login | `--name`, `--help`, `mudclient.toml`, a hidden password field | `clap` derive, `toml` + serde, `thiserror`, `ExitCode` |
| 22 | Go | Testing for real | `ParseCommand` that copes with every kind of whitespace | fuzzing (`go test -fuzz`), invariants, property-based tests, `go/parser` |

Proptest in Rust and a bot client are challenges in lesson 22.

**After phase 3** the world lives: things happen while you sit still, and it
remembers you the next day.

## Phase 4 – Tiles and zones (planned)

**Goal:** from room graph to tile world à la *Caves of Qud*: every room
becomes a **zone** – a map of e.g. 80×25 cells where you walk cell by cell.
Walk over the edge and you end up in the next zone. The room map from lesson
10 becomes the *world map*. The client gets real graphics.

| # | Track | Lesson | Contents |
|---|---|---|---|
| 23 | Go | Zones | Zones with tiles and entities with `(x, y)`. The protocol gets `zone` frames (the whole zone) and `delta` frames (only changes) – you learn why you don't send everything every tick. |
| 24 | Go | Procedural generation | Caves with cellular automata, tunnels with a "drunkard's walk", terrain with noise (simplex), all from a seed so the same world can be recreated. |
| 25 | Rust | The camera | Show part of a large zone, follow the player, scroll. Coordinate systems, `saturating_sub`, `clamp`. |
| 26 | Go | Field of view | Recursive *shadowcasting* on the server (so nobody can cheat) and "fog of war": cells you have seen but can't see now are drawn darker. |
| 27 | Go | Pathfinding | A* for creatures with `container/heap` and generics. Compare with Dijkstra maps (as in *Brogue*). |
| 28 | Rust | A graphical client | A window of its own with `macroquad` (simple) or `bevy` (large ecosystem). Load a CP437 tileset (16×16 glyphs – the same idea as Dwarf Fortress tilesets) or a CC0 pack like Kenney's *1-Bit Pack*. **A trait `Renderer` with two implementations** – terminal and tiles – and the rest of the client is unchanged. This is where the decision from lesson 10 pays off. |
| 29 | Rust | Light and life | Light sources with colour blending, animated water, particles, screen shake on hits. *Brogue*-style lighting. |
| 30 | Rust | The mouse | Click to walk (A* in the client), a mouse pointer showing "An old spruce" in a tooltip, a right-click menu. |

All client parts are then organized as a **Cargo workspace**: `mud-core`
(protocol, state), `mud-tui` and `mud-gfx`. A good exercise in modular design.

## Phase 5 – An extensible server (planned)

**Goal:** the server should never need a restart to get new content or new
features. This is the culmination of the course and ties everything together.

| # | Track | Lesson | Contents |
|---|---|---|---|
| 31 | Go | Reloading data | Watch the world files (polling or `fsnotify`), validate (lesson 09!), and swap the world *atomically* in the game loop. Players in removed rooms are moved to a safe place. The admin command `@reload`. |
| 32 | Go | Scripts in Lua | Embedded Lua with `gopher-lua` (pure Go). Commands and NPC behaviours in `.lua` files that reload live. A sandbox (no `os`/`io`), a time limit per script with `context`, and the hard question: *what API should scripts get?* The command registry from lesson 04 suddenly becomes very important. |
| 33 | Rust → Go | Plugins in WebAssembly | Write plugins in **Rust**, compile to `wasm32-unknown-unknown`, run them in the Go server with `wazero` (pure Go, no cgo). An ABI of your own: the plugin exports `on_command(ptr, len)`, the server offers `send(player, text)`. Memory limits, "fuel" against endless loops, versioning. **This is where the languages truly meet.** |
| 34 | Go | Events and hooks | An event bus: `OnEnterRoom`, `OnSay`, `OnTick`, `OnDeath`. Scripts and plugins subscribe; hooks can cancel actions ("the door is locked"). |
| 35 | Go | Copyover | The classic MUD technique from the 90s: the server starts a *new version of itself* with `exec` and hands over the open connections (file descriptors via `ExtraFiles`) and the player state. The players only see "The world shudders for a moment..." New code – without anyone being disconnected. |
| 36 | Go | Building in the game | *Online creation*: `@dig north Dark cave`, `@desc ...`, `@exit ...` – changes are validated, saved to the world file and visible to everyone immediately. |

Three levels of extensibility, from simplest to most powerful:

```
 data (JSON, live reloading)  →  scripts (Lua)  →  plugins (WASM)  →  a new binary (copyover)
     content                      behaviours        fast, isolated     everything else
                                                    code in any language
```

## Phase 6 – Out into the world (optional tracks)

Pick whatever interests you most:

- **Web client:** a WebSocket gateway in Go and the client compiled from Rust
  to WebAssembly – the same `render` code draws on the browser's `<canvas>`.
- **Async Rust:** rewrite the client with `tokio`. `tokio::select!` is
  surprisingly similar to Go's `select`. Build a load tester with 1000 bot players.
- **Real telnet:** IAC negotiation, NAWS (window size) and **GMCP**, so that
  ordinary MUD clients like Mudlet can show your map and status bar.
- **Operations:** `log/slog`, `pprof`, metrics for Prometheus, Docker, TLS,
  protection against spam and overload.
- **Internationalization for real:** every player picks a language, and the
  server chooses texts per player (a message catalog with `text/template`).
- **Rust on the server:** write the field-of-view or generator code as a Rust
  library and compare performance and developer experience.

---

## Concept map: Go and Rust side by side

| Idea | Go | Rust | Lessons |
|---|---|---|---|
| Compound data | `struct` | `struct` | 01, 03 |
| Methods | `func (s *Server) Serve()` | `impl Server { fn serve(&self) }` | 01, 03 |
| Errors | `error`, `if err != nil` | `Result<T, E>`, `?` | 01, 03, 05, 09, 11 |
| Missing value | `nil`, `value, ok :=` | `Option<T>` | 03, 05, 11 |
| Memory | garbage collector | ownership and borrowing | 03, 06 |
| Concurrency | goroutines | OS threads (`thread::spawn`), later `tokio` | 02, 06, (phase 6) |
| Messages | `chan T`, `select` | `mpsc::channel`, `recv`/`try_recv` | 06, 07, 13 |
| Shared data | `sync.Mutex` (voluntary) | `Mutex<T>` (required by the compiler) | 05, (14) |
| Abstraction | interfaces (implicit) | traits (explicit) | 06, 09, 10, 11, 13 |
| Generics | `func F[T any]` | `fn f<T: Trait>` | 06, (15, 27) |
| Variants | constants + `switch` | `enum` + `match` | 10, 11 |
| Text formatting | `fmt.Stringer` | `Display` | 11 |
| JSON | `encoding/json` + tags | `serde` + attributes | 09, 10, 11 |
| Tests | `go test`, table-driven | `cargo test`, `tests/` | all |
| Dependencies | `go get`, `go.mod` | `cargo add`, `Cargo.toml` | 11, 13 |

---

## Further reading

The course assumes no prior knowledge of Go or Rust, but some programming
experience (variables, loops, functions in some language). How to install
everything is described in [SETUP.md](SETUP.md). Extra reading per lesson:

| Lesson | Go: [Tour of Go](https://go.dev/tour) / [Effective Go](https://go.dev/doc/effective_go) | Rust: [The Rust Book](https://doc.rust-lang.org/book/) |
|---|---|---|
| 01 | Basics, Methods, Errors | – |
| 02 | Concurrency: Goroutines | – |
| 03 | – | ch. 1–5 (getting started, ownership, structs) and 9 (error handling) |
| 04 | More types: slices, maps, function values | – |
| 05 | Effective Go: Errors; the `sync` package | – |
| 06 | – | ch. 10 (generics, traits) and 16 (threads, channels) |
| 07 | Concurrency: Channels, Select | – |
| 08 | Pointers, Structs | – |
| 09 | `go doc encoding/json`, `go doc embed` | – |
| 10 | Interfaces; `go doc sync/atomic` | – |
| 11 | – | ch. 6 (enums, match), 7 (modules), [serde.rs](https://serde.rs) |
| 12 | – | ch. 13 (closures, iterators) |
| 13 | – | [ratatui.rs](https://ratatui.rs) – tutorials |

Good references when you get stuck: `go doc <package>`, `rustup doc --std`,
[Go by Example](https://gobyexample.com), [Rust by Example](https://doc.rust-lang.org/rust-by-example/),
and the compiler's error messages – especially Rust's, which often suggest the fix.

## Time required

Approximate times for someone who can program but is new to Go and Rust:

| Part | Time |
|---|---|
| Phase 1 (01–05) | 8–12 hours |
| Phase 2 (06–13) | 15–25 hours |
| Phase 3 (14–22) | 25–35 hours |
| Phase 4 | 30–45 hours |
| Phase 5 | 25–40 hours |

The Rust lessons typically take longer than the Go lessons. That is normal –
the borrow checker is a threshold everyone has to get over.
