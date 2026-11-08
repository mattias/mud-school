# Ledtrådar

## apply: tiles

```rust
for t in &d.tiles {
    let Some(code) = t.code.chars().next() else { continue };
    if let Some(cell) = self.rows.get_mut(t.y).and_then(|row| row.get_mut(t.x)) {
        *cell = code;
    }
    if !t.meaning.is_empty() {
        self.legend.insert(t.code.clone(), t.meaning.clone());
    }
}
```

## build_scene: zondelen

```rust
let (w, h) = (width as usize, zone_rows(height) as usize);
let (sx, sy) = zone.self_pos().unwrap_or((0, 0));
let (x0, y0) = (camera(sx, w, zone.width), camera(sy, h, zone.height));
let grid = render_zone(zone, x0, y0, w.min(zone.width), h.min(zone.height));
for (vy, row) in grid.iter().enumerate() {
    for (vx, c) in row.iter().enumerate() {
        let (zx, zy) = (x0 + vx, y0 + vy);
        // skip ' ' (never seen); otherwise push a SceneCell at (vx, vy)
    }
}
```

## De sista N loggraderna

```rust
app.log.iter().rev().take(LOG_LINES as usize).rev().cloned().collect()
```

`rev()` två gånger: ta från slutet, och lägg dem sedan i ordning igen.

## "could not determine which binary to run"

Du har två program nu. `default-run = "mud-client"` under `[package]`
säger till `cargo run` vilket du menar när du inte säger det.
