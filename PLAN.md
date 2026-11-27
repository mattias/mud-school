# PLAN.md – where MUD School stands

## Status

All six planned phases are implemented, verified and merged: **44 lessons**,
each in English, Swedish and Japanese, with tests and reference solutions.
`./mudschool verify` passes for all 132 lesson/language pairs.

| Phase | Lessons | Content | PR |
|---|---|---|---|
| 1 The basics | 01–05 | TCP server, goroutines, commands, players; first Rust client | #1 |
| 2 A world and a face | 06–13 | chat, rooms, world files, protocol + map; threads, serde, ratatui TUI | #1 (fixes #2–#4) |
| 3 A living world | 14–22 | game loop (actor), items, NPCs, combat, SQLite saving, settings, fuzzing | #5 |
| 4 Tiles and zones | 23–30 | zones, generation, camera, FOV, A*, macroquad window, light, mouse | #8 |
| 5 An extensible server | 31–36 | live reload, Lua, WASM plugins (Rust), events/hooks, copyover, OLC | #9 |
| 6 Out into the world | 37–44 | tutorial, languages per player, WebSocket, Rust core → WASM in the browser, Docker + Caddy on a VPS, protection, tokio load tester | #10 |

The details of every lesson are in [docs/en/CURRICULUM.md](docs/en/CURRICULUM.md).

## Decisions already made

- **Two languages, on purpose:** Go for the server, Rust for the client. They
  meet in lesson 33 (Rust plugins in the Go server) and 41 (Rust core in the browser).
- **Tests are the specification.** Students write code until the lesson's
  tests pass; the tests use the message constants, so every language works.
- **The course stays a CLI** (`./mudschool`). The *game* goes on the web
  (lessons 39–42), not the course.
- **Deployment target:** an ordinary VPS with Docker and Caddy (lesson 42).
- **Only the world is translated per player** (lesson 38). The server's own
  messages stay in the course language; per-player messages are a challenge.
- **Copyover is off while the web client is on** (lesson 39): WebSockets can't
  be handed over. Documented, with a challenge.
- **Known slow spot left on purpose:** `syncAll` is O(n²) for many players in
  one room. Lesson 44 measures it and leaves the fix as the last challenge.

## Possible next steps

Nothing is scheduled. If the course grows, these are the candidates, roughly
in order of how well they fit (more in [docs/en/IDEAS.md](docs/en/IDEAS.md)):

1. **Polish what exists** – work through [TODO.md](TODO.md); have someone take
   the later phases on Windows end to end (the user has tried the first lessons there).
2. **A phase 7 "scaling the world"**: fix `syncAll` with per-room player lists
   and dirty rooms, measured with the lesson 44 load tester; several servers,
   one world (sharding).
3. **Real telnet**: IAC negotiation, NAWS and GMCP so Mudlet can show the map.
4. **Server messages per player**: a message catalog per language,
   `p.Msg(...)` – a big refactor across every lesson, so it would be a new
   phase or a separate track, not edits to old lessons.
5. **Rust on the server**: FOV or zone generation as a Rust library run by
   wazero (the lesson 33 ABI) – compare speed.
6. **An async terminal client** with tokio, replacing threads and channels.

Before starting any of these: agree the scope with the user, write it into
this file, and follow the workflow in [CLAUDE.md](CLAUDE.md).
