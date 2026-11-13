# Lektion 41 – Klienten i webbläsaren

Sedan lektion 39 kan du spela i webbläsaren – med text. Nu får webbläsaren
kartan: kärnan från lektion 40, kompilerad till **WebAssembly**, ritar zonen
på en `<canvas>`, med facklors ljus och dimma, och du klickar för att gå –
samma A* som i fönstret.

Sidan (`server/web/mud.js`, given sedan lektion 39) vet redan hur den ska
prata med den: finns `mud_web.wasm` laddar den in den.

```
WebSocket ──rad──▶ mud.js ──on_line──▶ mud_web.wasm  (zon, deltan, status)
                   mud.js ◀──frame──── mud_web.wasm  (cols × rows rutor, 60 gånger i sekunden)
klick     ──col,row──────────▶ mud_web.wasm ──mud.send("east")──▶ mud.js ──▶ WebSocket
```

## Mål

- Kompilera Rust till WebAssembly: `--target wasm32-unknown-unknown` och en `cdylib`.
- Utforma ett litet **ABI** mellan två språk – tal och delat minne – och se
  att det är lektion 33 igen, åt andra hållet.
- Håll logiken testbar: `WebClient` är vanlig Rust; bara ett tunt lager vet
  om WebAssembly.
- Packa data åt ett annat språk: byte i en känd ordning.

## Rust-koncept

### En `cdylib`

```toml
[lib]
crate-type = ["cdylib", "rlib"]
```

`cdylib` är det WebAssembly behöver: ett bibliotek med ett C-likt gränssnitt.
`rlib` är ett vanligt Rust-bibliotek – för testerna, som använder `WebClient` direkt.

### ABI:t, version 1

I lektion 33 var Go-servern värd och ett Rust-plugin gäst. Nu är webbläsaren värd:

| exporter | |
|---|---|
| `abi_version() -> i32` | 1 |
| `alloc(len) -> ptr`, `dealloc(ptr, len)` | minne som sidan skriver rader i |
| `on_line(ptr, len)` | en rad från servern |
| `frame(cols, rows, t_ms) -> ptr` | skärmen: cols × rows rutor à 12 byte |
| `click(col, row)`, `tick(t_ms)` | ett klick; tiden går |

och den importerar `mud.send(ptr, len)` – ett kommando till servern. En ruta
är 12 byte: tecknet (`u32`, little endian), förgrundens r g b, bakgrundens
r g b, 1 om det finns en bakgrund, och en reservbyte.

### `#[cfg(target_arch = "wasm32")]` och `#[no_mangle]`

De exporterade funktionerna finns bara när man kompilerar för WebAssembly.
De håller en `WebClient` i en `thread_local!` (webbläsaren har en tråd) och
anropar dess metoder:

```rust
#[no_mangle]
pub extern "C" fn frame(cols: usize, rows: usize, t_ms: f64) -> *const u8 {
    CLIENT.with_borrow_mut(|c| c.frame(cols, rows, t_ms as u64).as_ptr())
}
```

`#[no_mangle]` behåller namnet `frame` så att sidan hittar det. Byten ligger
kvar i `WebClient` efter anropet, så pekaren är fortfarande giltig när sidan läser dem.

## Kontraktet

- `client/web/` är ett paket `mud-web` (givet: `Cargo.toml`), en medlem i
  workspacen och ett dev-beroende för klienten (så att testerna kan använda det).
- `WebClient::new()`, `on_line(&str)` (zon, delta och status sparas; en zon
  för ett annat rum avbryter en promenad; allt annat ignoreras),
  `frame(cols, rows, t_ms) -> &[u8]`, `click(col, row)`, `tick(t_ms)` och
  `take_outbox() -> Vec<String>`. `ABI_VERSION = 1`, `CELL_BYTES = 12`, `WALK_MS`.
- `frame`: `cols × rows × CELL_BYTES` byte, alla noll utan zon. Med en zon:
  kameran från lektion 25 runt spelaren; en zon som är mindre än bilden
  centreras; varje ruta från `render_zone`, dess färger `shade`ade med
  `zone_light` (terrängen från senaste status).
- `click(col, row)`: zonrutan under klicket (som den ritades i senaste
  bilden), en väg med `find_path`, stegen som kommandon; `tick` skickar högst
  ett per `WALK_MS`.
- Byggd för `wasm32-unknown-unknown` exporterar `mud_web.wasm` `memory`,
  `abi_version`, `alloc`, `dealloc`, `on_line`, `frame`, `click`, `tick`, och
  importerar bara `mud.send`.

## Uppgift

1. `rustup target add wasm32-unknown-unknown` (om du inte gjort det sedan lektion 33).
2. Lägg till `"web"` bland workspacens medlemmar och `mud-web = { path = "web" }`
   under `[dev-dependencies]` i `client/Cargo.toml`.
3. `web/src/lib.rs` är delvis given (konstanterna, structen, `take_outbox`
   och ABI-lagret): skriv `on_line`, `frame`, `click` och `tick`.
4. `./mudschool run web` bygger `mud_web.wasm`, kopierar den till `server/web/`
   och startar servern. Öppna http://127.0.0.1:8080.

## Go ↔ Rust

| | lektion 33 | lektion 41 |
|---|---|---|
| värd | Go-servern (wazero) | webbläsaren (mud.js) |
| gäst | ett Rust-plugin | Rust-klienten |
| in till gästen | `"player\tverb\targs"` | en rad från servern |
| ut från gästen | `mud.send` – text till en spelare | `mud.send` – ett kommando; `frame` – blivande pixlar |

## Fundera på

- Varför returnerar `frame` en pekare in i `WebClient` i stället för en ny `Vec`?
  Vad skulle gå fel med en `Vec` som släpps i slutet av funktionen?
- `CELL_BYTES` är 12, inte 11. Varför kan en multipel av 4 vara trevligare för JavaScript?
- Sidan ritar 60 bilder i sekunden, men zonen ändras några gånger i sekunden.
  Hur skulle du kunna slippa jobbet?

## Utmaningar

- Rita tiles i stället för tecken: servera `tiles.png` (lektion 28) och låt
  `frame` skicka med spritenumren också.
- Hover: en tooltip med `describe_cell` (lektion 30 – flytta den till kärnan först).
- Partiklar i webbläsaren (lektion 29).
