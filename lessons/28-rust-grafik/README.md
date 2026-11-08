# Lesson 28 – A graphical client

In lesson 10 the server learned to send *meaning* – `"forest"`, `"wall"`,
`"creature"` – instead of appearance. The terminal client turned meanings
into coloured characters. Now you add a second client: a **window** of its
own, with pictures – and the rest of the client (the protocol, the network,
the App with its keys and history) stays **exactly** the same.

```
the server  ──JSON──>  protocol.rs  ──>  App  ──>  build_scene  ──>  Scene
                                                                      │
                                    ┌─────────────── Renderer ────────┴────────┐
                                    │                    │                     │
                              TextRenderer          the terminal          GfxRenderer
                              (tests)               (ratatui)             (a window, tiles)
```

The window uses **macroquad**, a small game library, and a **tileset**: one
PNG with all the 16×16 pictures side by side (an *atlas*). The course draws
its own (`client/assets/tiles.png`, CC0, made by `tools/tileset/main.go`) –
the same idea as the tilesets for *Dwarf Fortress*. Swap it for another if you like.

## Goals

- Design a **trait** that hides *how* something is drawn – and implement it more than once.
- Separate *what* is on the screen (a `Scene`, plain data) from *how* it is
  drawn. Plain data is easy to test.
- Use **Cargo features** and optional dependencies, and a second binary.
- Draw sprites from an atlas: index → rectangle in the PNG.

## Rust concepts

### A trait with several implementations

```rust
pub trait Renderer {
    fn cell(&mut self, x: u16, y: u16, cell: &SceneCell);
    fn text(&mut self, x: u16, y: u16, text: &str, fg: Rgb);
}

pub fn draw_scene(scene: &Scene, r: &mut impl Renderer) { ... }
```

`draw_scene` doesn't know if it draws characters into a grid or textures
into a window. `impl Renderer` in the argument is *static dispatch*: the
compiler makes one copy of `draw_scene` per renderer type.

### Cargo features

The graphics library is big, and the terminal client doesn't need it:

```toml
macroquad = { version = "0.4", optional = true }

[features]
gfx = ["dep:macroquad"]

[[bin]]
name = "mud-gfx"
path = "gfx/main.rs"
required-features = ["gfx"]
```

The program is in `gfx/main.rs`, not in `src/bin/`: Cargo finds programs in
`src/bin/` by itself – and would try to build this one before you have added
macroquad. The `[[bin]]` entry tells Cargo where it is.

`cargo test` and `cargo run` don't build macroquad at all. `cargo run
--release --features gfx --bin mud-gfx` does – and in the code,
`#[cfg(feature = "gfx")] pub mod gfx;` only exists then. Add
`default-run = "mud-client"` under `[package]` so that `cargo run` still
knows which program you mean.

### The fog of war, at last

Since lesson 26 deltas have `tiles` (newly seen tiles) and `visible` (the
field of view) – and your client has ignored them. Now `ZoneView::apply` puts
the tiles into the zone and keeps `visible`; cells you have seen but don't see
right now are drawn darker (`shade(colour, FOG)`), in the window *and* in the
terminal.

## The contract

- `protocol.rs`: `ZoneData::visible: Vec<String>`; `TileChange {x, y, code,
  meaning}`; `DeltaData::tiles` and `DeltaData::visible` (missing = empty).
- `zone.rs`: `ZoneView::visible: Vec<Vec<bool>>` (from the zone frame),
  `is_visible(x, y)` (true everywhere when the server sent no `visible`);
  `apply` puts `tiles` into `rows` and their meanings into `legend`, and
  replaces `visible` when the delta has one.
- `tileset.rs`: `TILE_PX = 16.0`, `ATLAS_COLUMNS = 16`; `tile_sprite(meaning)`:
  floor 0, wall 1, grass 2, tree 3, water 4, rock 5, door 6, stairs_up 7,
  stairs_down 8, otherwise `None`; `entity_sprite(kind, is_self)`: yourself 9,
  player 10, creature 11, item 12, otherwise 13; `atlas_rect(index)` → `(x, y,
  w, h)` in pixels (16 tiles per row).
- `scene.rs`: `SceneCell {x, y, glyph, fg, bg, tile, entity, light}`, `Scene
  {width, height, cells, status, log, input}`, `LOG_LINES = 5`, `FULL_LIGHT`,
  `FOG`, `shade(colour, light)` (each component × light / 255),
  `zone_rows(height)` = `height - 2 - LOG_LINES`, and:
  - `build_scene(app, width, height)` – `status` (the status bar's text, or
    `NOT_LOGGED_IN` trimmed), the last `LOG_LINES` log lines, `input` = `"> "`
    + the input as shown; with a zone: one cell per **explored** cell in the
    camera's view (the camera from lesson 25, `width × zone_rows(height)`), at
    view coordinates, with the glyph and colours from `render_zone`, `tile` and
    `entity` from the tileset, and `light` = `FULL_LIGHT` if visible, else `FOG`;
  - `draw_scene(scene, renderer)` – the status on row 0, the cells one row
    down (`y + 1`), the log under the zone, the input on the last row;
  - `TextRenderer` – draws into a grid of characters; `lines()`.
- `gfx.rs` and `gfx/main.rs` (given): the window renderer and the program.
- `Cargo.toml`: the feature, the optional dependency, the second binary,
  `default-run`. `assets/tiles.png` is given.
- The terminal client dims explored-but-not-visible cells too.

## Task

1. The protocol and `zone.rs` (tiles, visible).
2. `tileset.rs` (`atlas_rect`) and `scene.rs`.
3. `Cargo.toml`. Then run it: `./mudschool run server` and
   `./mudschool run gfx --name Anna`. The first build takes a while.

On Linux you may need the development packages for X11 and OpenGL (Ubuntu:
`sudo apt install libx11-dev libxi-dev libgl1-mesa-dev`); Windows and macOS
need nothing extra.

The window uses macroquad's built-in font, which only has Latin characters:
Japanese (and å, ä, ö) in the log become boxes. See the challenges.

## Think about

- Why is a `Scene` (plain data) easier to test than code that draws directly?
- `&mut impl Renderer` vs. `&mut dyn Renderer`: what is the difference? When
  would you need `dyn`?
- The window redraws everything 60 times a second; the terminal only when
  something happens. Why is that fine for a game?
- What would it take to show *both* the terminal and the window at once?

## Challenges

- A font with Japanese and Swedish characters: `load_ttf_font` and a `font =
  "..."` setting in `mudclient.toml` (e.g. *Noto Sans JP* or *UDEV Gothic*).
- Zoom with the mouse wheel (change `scale`).
- Another tileset: Kenney's *1-Bit Pack* (CC0) – write a new `tile_sprite`
  for it, or let `mudclient.toml` choose the atlas file.
- Organize the client as a **Cargo workspace**: `mud-core` (protocol, App,
  scene), `mud-tui` and `mud-gfx`.
