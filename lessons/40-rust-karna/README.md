# Lesson 40 – The core

Lesson 41 will run your Rust client in the browser. That sounds like a
rewrite – but most of the client doesn't care where it runs. Parsing a line
from the server, keeping the zone up to date, choosing a glyph and a colour
for a forest, the light of a torch, a path to a clicked cell: none of it
needs a terminal, a window, a network or a clock.

So the client is split in two. The **core** – a library crate of its own,
`mud-core` – gets everything that is the same everywhere. The terminal
client, the window and (next lesson) the browser are thin shells around it:

```
client/                      a Cargo *workspace*
├── Cargo.toml               mud-client: the terminal and the window
├── src/                     app, scene, net, config, messages, gfx …
└── core/                    mud-core: protocol, zone, render, tileset,
    ├── Cargo.toml                     light, path – no I/O
    └── src/
```

## Goals

- Split a crate into a **workspace** with a library of its own.
- Draw the line between pure logic and I/O – and keep it with a test.
- Meet Rust's **orphan rule**, and work around it.
- Inject time instead of reading a clock.

## Rust concepts

### A workspace

```toml
[dependencies]
mud-core = { path = "core" }

[workspace]
members = ["core"]
```

The members share one `Cargo.lock` and one `target/`. `mud-client` depends
on `mud-core` like on any crate, by path.

### Re-exports keep the old paths

```rust
pub use mud_core::{protocol, render, tileset, zone};
```

in the client's `lib.rs` means `crate::protocol::ServerMsg` and
`mud_client::zone::ZoneView` still work – the same types, now living in the core.

### The orphan rule

Two things stop compiling after the move:

```rust
impl fmt::Display for Status { ... }   // Display: std's.  Status: the core's.
impl From<Rgb> for Color { ... }       // From: std's.  Rgb: the core's.  Color: ratatui's.
```

Rust only lets you implement a trait for a type if your crate owns the trait
*or* the type. Otherwise two crates could implement the same thing in two
ways, and nobody would know which one to use. The client owns neither, so
these become plain functions: `status_line(&Status)` and `color(Rgb)`. (The
core *could* implement `Display` – but the word "online" is the client's, in
the course language.)

### Time as a number

The core may not call `Instant::now()` – in the browser it panics. Light
that shimmers gets the time from outside, as milliseconds:

```rust
pub fn zone_light(zone: &ZoneView, terrain: &str, x: usize, y: usize, t_ms: u64) -> Rgb
```

The terminal client passes the time since it started; the browser will pass
`requestAnimationFrame`'s clock; the tests pass whatever they like.

## The contract

- `client/core/` is a package `mud-core` (given: `Cargo.toml` and `src/lib.rs`)
  that depends only on `serde` and `serde_json`, and whose code does no I/O
  and reads no clock (no `std::net`, `std::fs`, `std::thread`, `Instant::now`, …).
- It has the modules `protocol`, `zone`, `render`, `tileset` (moved as they
  are, except `Display for Status`), `light` (the pure part: `Light`, `TORCH`,
  `ambient`, `light_at`, `lights_for`, `shimmer`, `scale` – and now `shade`,
  `FOG`, `FULL_LIGHT` and the new `zone_light`) and `path` (`Cell`, `walkable`,
  `STEPS`, `step_command`, `find_path` from `mouse.rs`).
- `zone_light(zone, terrain, x, y, t_ms)`: `FOG` if the cell isn't visible,
  otherwise the ambient light plus the torches, and water shimmers with `t_ms`.
- The client depends on `mud-core`, re-exports `protocol`, `render`, `tileset`
  and `zone`, and its `light` and `mouse` re-export the core's parts – the old
  paths still work.
- `scene::status_line(&Status) -> String` (with `messages::ONLINE`) and
  `scene::cell_light` uses `zone_light`; `app::color(Rgb) -> Color`.

## Task

1. Add `mud-core = { path = "core" }` and `[workspace] members = ["core"]` to
   `client/Cargo.toml`.
2. Move `protocol.rs`, `zone.rs`, `render.rs` and `tileset.rs` to `core/src/`
   (`git mv` if you use git), and take `impl Display for Status` out.
3. Split `light.rs` (particles and the shake stay) and `mouse.rs` (the menu
   and the tooltip stay; the paths move to `core/src/path.rs`).
4. Write `zone_light`, and fix what the compiler finds: `crate::` paths in the
   core, `status_line`, `color`.

Run `./mudschool run client` and `./mudschool run gfx`: everything should be as before.

## Go ↔ Rust

| | Go | Rust |
|---|---|---|
| Several packages, one build | a module with packages in folders | a workspace with crates |
| Who may add methods | only the package that defines the type | the crate that owns the trait or the type (orphan rule) |
| Re-export | – (callers import the new package) | `pub use` |

## Think about

- Which modules could move to the core, but didn't? Why not `messages.rs`?
- The orphan rule forbids `impl From<Rgb> for Color` in the client. Where
  *could* it live?
- Why is "no clock" more than a WebAssembly detail? (Think of the tests.)

## Challenges

- Move the parts of `App` that aren't about the terminal into the core: a
  `GameState` that both the terminal and the browser could use.
- `cargo build --target wasm32-unknown-unknown -p mud-core` – does it build?
  It should, already now.
