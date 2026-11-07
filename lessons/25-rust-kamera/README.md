# Lesson 25 – The camera

Since lesson 23 the server sends zones and deltas – and your client has
ignored them politely (`#[serde(other)] Unknown`, lesson 11). Now it draws
them: the zone fills most of the screen, the log moves down, and the arrow
keys walk.

A zone is 31×13 cells by default but can be much bigger than the terminal.
Then the client shows a *window* of it – a **camera** that follows the player
and never shows what lies outside the zone.

```
zone (40 wide)   ##########################################
camera (10)                 [ ....@..... ]       focus 20 -> starts at 15
near the edge                               [ .......@.# ]  never past the end
```

## Goals

- Deserialize nested JSON with lists of structs, and a field called `self`
  (a keyword in Rust) with `#[serde(rename = "self")]`.
- Keep state up to date with incremental updates (deltas).
- Use `BTreeMap` when the order matters and `HashMap` when it doesn't.
- Calculate safely with unsigned numbers: `saturating_sub`, `min`.
- Lay out a more complex screen with nested ratatui layouts.

## Rust concepts

### `self` as a field name

```rust
#[derive(Deserialize)]
pub struct ZoneData {
    // ...
    #[serde(rename = "self", default)]
    pub self_id: String,
}
```

### `From`

`impl From<ZoneData> for ZoneView` converts what the server sent into what
the client keeps: the rows as `Vec<Vec<char>>` (so `rows[y][x]` is a
character, not a byte position in a UTF-8 string) and the entities in a map
by ID.

### `saturating_sub`

`usize` can't be negative: `2 - 5` panics in a debug build.
`2usize.saturating_sub(5)` is `0`. That is exactly what a camera at the left
edge needs:

```rust
focus.saturating_sub(view / 2).min(size - view)
```

### Match guards

The arrow keys do two things now – walk with an empty input line, browse the
history while you type:

```rust
KeyCode::Up | KeyCode::Down | KeyCode::Left | KeyCode::Right
    if self.input.is_empty() && self.history_pos.is_none() && !self.secret =>
{
    return walk_command(key.code).map(String::from);
}
KeyCode::Up if !self.secret => self.history_up(),
```

## The contract

- `protocol.rs`: `Entity {id, kind, name, glyph, x, y}`, `ZoneData {room,
  width, height, rows, legend, entities, self_id}` and `DeltaData {room,
  added, moved, removed}` (missing lists = empty); `ServerMsg::Zone { zone }`
  and `ServerMsg::Delta { delta }`.
- `zone.rs` (partly given): `ZoneView` with `From<ZoneData>`, `apply(delta) ->
  bool` (a delta for another room is ignored and returns `false`),
  `self_pos()`, `code_at`, `meaning_at`, `entity_at`; and
  `camera(focus, view, size) -> usize`: `focus` in the middle, clamped to
  `0..=size-view`, and `0` when the zone fits.
- `render.rs`: tiles for `floor`, `wall`, `grass`, `tree`, `rock`, `door`;
  `entity_cell(kind, glyph, is_self)`; `render_zone(view, x0, y0, w, h)`.
- `app.rs`: `App::zone: Option<ZoneView>`, set by zone frames and updated by
  deltas; with a zone the screen is the zone on top and the log below; with an
  empty input line the arrow keys return `"north"`, `"south"`, `"west"`,
  `"east"` (and nothing goes into the log or the history); while typing,
  Up/Down browse the history as before.

## Task

Protocol first, then `zone.rs`, `render.rs` and the app. Run the server from
lesson 24 and walk around with the arrow keys. Make the terminal small and
watch the camera.

## Think about

- Why can a delta arrive for "the wrong" room? (Think about the order of frames
  when you walk out through an opening.)
- `entities` is a `BTreeMap`, `legend` a `HashMap`. Why the difference?
- `render_zone` asks `entity_at` for every cell – it searches all entities.
  How many cells and entities before that matters? How would you speed it up?
- What happens if the server sends a delta before any zone at all?

## Challenges

- Diagonal steps: the number keys 7, 9, 1, 3 (or `y u b n` like in *NetHack*).
- A border of a different colour where the zone ends, when the zone is smaller than the screen.
- Hold Shift + arrow to keep walking until something interesting happens ("run").
