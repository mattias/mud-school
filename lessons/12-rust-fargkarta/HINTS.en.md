# Hints

## tile

```rust
pub fn tile(meaning: &str) -> Cell {
    match meaning {
        "forest" => Cell { glyph: '♣', fg: Rgb(40, 170, 70), bg: None },
        // ...
        _ => Cell { glyph: '?', fg: Rgb(255, 0, 255), bg: None },
    }
}
```

A small helper function `const fn cell(glyph: char, fg: Rgb) -> Cell` saves a
lot of typing.

## render_map

```rust
map.rows
    .iter()
    .map(|row| {
        row.chars()
            .map(|code| match code {
                ' ' => BLANK,
                _ => match map.meaning(code) {
                    Some(meaning) => tile(meaning),
                    None => /* show code in grey */,
                },
            })
            .collect()
    })
    .collect()
```

## to_ansi

```rust
for (i, row) in grid.iter().enumerate() {
    if i > 0 {
        out.push('\n');
    }
    for c in row {
        let Rgb(r, g, b) = c.fg;
        write!(out, "\x1b[38;2;{r};{g};{b}m").unwrap();
        // background if c.bg is Some(...)
        out.push(c.glyph);
        out.push_str("\x1b[0m");
    }
}
```
