# Lesson 12 – The map in colour

Now it gets pretty. Dwarf Fortress, Caves of Qud, NetHack and Brogue show
whole worlds with *one character per cell* – but with well-chosen characters
and colours it becomes incredibly atmospheric. We do the same: the map's
codes become coloured **tiles**.

```
    Ω─Ω                  ♣ green forest, ≈ blue water (with a dark blue background),
    │ │                  ▲ grey mountain, # stone houses, · road, @ you (yellow)
    ▲ Ω
    │
≈─"─·─♣─♣
      │ │
      · ♣
```

This is also the first step towards real tile graphics in phase 4: there we
replace `Cell { glyph: '♣', ... }` with "tile number 5 in the sprite sheet" –
the rest of the code stays the same.

## Goals

- Create small **Copy types** (`Rgb`, `Cell`) that are copied instead of moved.
- Use `match` on strings with a fallback arm `_`.
- Use **iterators**: `iter`, `chars`, `map`, `collect`, `enumerate`.
- Understand how terminals draw colour with **ANSI escape codes**.

## Rust concepts

### Copy

```rust
#[derive(Clone, Copy, PartialEq)]
pub struct Rgb(pub u8, pub u8, pub u8);   // a "tuple struct": fields without names

let a = Rgb(255, 0, 0);
let b = a;          // a COPY – a can still be used (thanks to Copy)
let Rgb(r, g, b) = a;   // destructuring
```

Without `Copy`, `let b = a;` would have *moved* the value. Small values made
only of numbers are perfect as `Copy`. `String` and `Vec` can never be.

### Iterators

```rust
let grid: Vec<Vec<Cell>> = map.rows
    .iter()                                  // &String for each row
    .map(|row| row.chars().map(to_cell).collect())  // Vec<Cell> per row
    .collect();                              // Vec<Vec<Cell>>
```

Iterators are *lazy*: nothing happens until `collect` (or a `for` loop) pulls
on them. The compiler turns the chain into code as fast as a hand-written loop.

**Go ↔ Rust:** in Go you almost always write `for` loops. In Rust both styles
are common – pick whichever is clearest.

### ANSI codes

Terminals interpret certain character sequences as commands:

| Code | Meaning |
|---|---|
| `\x1b[38;2;R;G;Bm` | foreground colour (truecolor) |
| `\x1b[48;2;R;G;Bm` | background colour |
| `\x1b[0m` | reset |

Try: `printf '\x1b[38;2;40;170;70m♣\x1b[0m\n'`

Every terminal library (crossterm, ratatui, ncurses) does this under the hood.
Now you do it yourself once – in the next lesson we let a library do it.

## Task

1. **`src/render.rs`** (new file): write `tile`, `render_map` and `to_ansi`.
   Add `pub mod render;` to `lib.rs`.
   - `tile` gives every meaning its own character and a fitting colour (the
     tests check e.g. that forest is green and water blue). Unknown meanings
     give a clear fallback cell, e.g. a magenta `?`.
   - `render_map`: spaces become `BLANK`; a code in the legend becomes
     `tile(meaning)`; a code that is *not* in it is shown as itself in grey.
   - `to_ansi` as the comment in the file says.
2. **`src/main.rs`**: draw the map with `to_ansi(&render_map(&map))`.

## Try it yourself

Run the server and the client and walk around: `s`, `s`, `s`, `s`, `e` ...

> Seeing strange characters? Your terminal must use UTF-8 and a font that has
> the characters. No colours? The terminal may lack truecolor support (most
> modern ones have it). See `docs/en/SETUP.md`.

## Think about

- Why does `to_ansi` take a `&[Vec<Cell>]` and not a `Vec<Vec<Cell>>`?
- `tile` returns a new `Cell` every time. Why is that no problem for performance?

## Challenges

- Vary the forest: let `render_map` choose between `♣`, `♠` and `↑` based on
  the position (`enumerate` gives you x and y), so that the forest looks
  alive. It is a common trick in roguelikes.
- Read about *colour theory for roguelikes*: darker colours for things further
  away, a palette with a limited number of tones (e.g. the "PICO-8" palette).
