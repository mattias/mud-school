# Ledtrådar

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

En liten hjälpfunktion `const fn cell(glyph: char, fg: Rgb) -> Cell` sparar
mycket skrivande.

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
                    None => /* visa code i grått */,
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
        // bakgrund om c.bg är Some(...)
        out.push(c.glyph);
        out.push_str("\x1b[0m");
    }
}
```
