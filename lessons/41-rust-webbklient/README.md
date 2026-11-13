# Lesson 41 – The client in the browser

Since lesson 39 you can play in the browser – with text. Now the browser
gets the map: the core from lesson 40, compiled to **WebAssembly**, draws the
zone on a `<canvas>`, with torchlight and fog, and you click to walk – the
same A* as in the window.

The page (`server/web/mud.js`, given since lesson 39) already knows how to
talk to it: if `mud_web.wasm` is there, it loads it.

```
WebSocket ──line──▶ mud.js ──on_line──▶ mud_web.wasm  (zone, deltas, status)
                    mud.js ◀──frame──── mud_web.wasm  (cols × rows cells, 60 times a second)
click     ──col,row──────────▶ mud_web.wasm ──mud.send("east")──▶ mud.js ──▶ WebSocket
```

## Goals

- Compile Rust to WebAssembly: `--target wasm32-unknown-unknown` and a `cdylib`.
- Design a small **ABI** between two languages – numbers and shared memory –
  and see that it is lesson 33 again, the other way round.
- Keep the logic testable: the `WebClient` struct is ordinary Rust; only a thin
  layer knows about WebAssembly.
- Pack data for another language: bytes in a known layout.

## Rust concepts

### A `cdylib`

```toml
[lib]
crate-type = ["cdylib", "rlib"]
```

`cdylib` is what WebAssembly needs: a library with a C-like interface. `rlib`
is an ordinary Rust library – for the tests, which use `WebClient` directly.

### The ABI, version 1

In lesson 33 the Go server was the host and a Rust plugin the guest. Now the
browser is the host:

| exports | |
|---|---|
| `abi_version() -> i32` | 1 |
| `alloc(len) -> ptr`, `dealloc(ptr, len)` | memory for the page to write lines into |
| `on_line(ptr, len)` | a line from the server |
| `frame(cols, rows, t_ms) -> ptr` | the screen: cols × rows cells of 12 bytes |
| `click(col, row)`, `tick(t_ms)` | a click; time passes |

and it imports `mud.send(ptr, len)` – a command for the server. A cell is
12 bytes: the character (`u32`, little endian), the foreground r g b, the
background r g b, 1 if there is a background, and a spare byte.

### `#[cfg(target_arch = "wasm32")]` and `#[no_mangle]`

The exported functions only exist when compiling for WebAssembly. They keep
one `WebClient` in a `thread_local!` (the browser has one thread) and call
its methods:

```rust
#[no_mangle]
pub extern "C" fn frame(cols: usize, rows: usize, t_ms: f64) -> *const u8 {
    CLIENT.with_borrow_mut(|c| c.frame(cols, rows, t_ms as u64).as_ptr())
}
```

`#[no_mangle]` keeps the name `frame` so the page can find it. The bytes
stay in `WebClient` after the call, so the pointer is still good when the
page reads them.

## The contract

- `client/web/` is a package `mud-web` (given: `Cargo.toml`), a workspace
  member, and a dev-dependency of the client (so the tests can use it).
- `WebClient::new()`, `on_line(&str)` (zone, delta and status are kept; a
  zone for another room stops a walk; everything else is ignored),
  `frame(cols, rows, t_ms) -> &[u8]`, `click(col, row)`, `tick(t_ms)` and
  `take_outbox() -> Vec<String>`. `ABI_VERSION = 1`, `CELL_BYTES = 12`, `WALK_MS`.
- `frame`: `cols × rows × CELL_BYTES` bytes, all zero without a zone. With a
  zone: the camera from lesson 25 round the player; a zone smaller than the
  frame is centred; each cell from `render_zone`, its colours `shade`d with
  `zone_light` (the terrain from the last status).
- `click(col, row)`: the zone cell under it (as drawn in the last frame), a
  path with `find_path`, the steps as commands; `tick` sends one at most
  every `WALK_MS`.
- Built for `wasm32-unknown-unknown`, `mud_web.wasm` exports `memory`,
  `abi_version`, `alloc`, `dealloc`, `on_line`, `frame`, `click`, `tick`, and
  imports only `mud.send`.

## Task

1. `rustup target add wasm32-unknown-unknown` (if you haven't since lesson 33).
2. Add `"web"` to the workspace members and `mud-web = { path = "web" }` under
   `[dev-dependencies]` in `client/Cargo.toml`.
3. `web/src/lib.rs` is partly given (the constants, the struct, `take_outbox`
   and the ABI layer): write `on_line`, `frame`, `click` and `tick`.
4. `./mudschool run web` builds `mud_web.wasm`, copies it to `server/web/` and
   starts the server. Open http://127.0.0.1:8080.

## Go ↔ Rust

| | lesson 33 | lesson 41 |
|---|---|---|
| host | the Go server (wazero) | the browser (mud.js) |
| guest | a Rust plugin | the Rust client |
| into the guest | `"player\tverb\targs"` | a line from the server |
| out of the guest | `mud.send` – text to a player | `mud.send` – a command; `frame` – pixels to be |

## Think about

- Why does `frame` return a pointer into `WebClient` instead of a fresh `Vec`?
  What would go wrong with a `Vec` that is dropped at the end of the function?
- `CELL_BYTES` is 12, not 11. Why might a multiple of 4 be nicer for JavaScript?
- The page draws 60 frames a second, but the zone changes a few times a
  second. How could you avoid the work?

## Challenges

- Draw tiles instead of glyphs: serve `tiles.png` (lesson 28) and give
  `frame` the sprite numbers too.
- Hover: a tooltip with `describe_cell` (lesson 30 – move it to the core first).
- Particles in the browser (lesson 29).
