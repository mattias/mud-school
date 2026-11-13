# Hints

## on_line

```rust
match parse_line(line) {
    ServerMsg::Zone { zone } => {
        if self.zone.as_ref().is_some_and(|z| z.room != zone.room) {
            self.walk.clear();
        }
        self.zone = Some(ZoneView::from(zone));
    }
    ServerMsg::Delta { delta } => { /* apply */ }
    ServerMsg::Status { status } => self.status = Some(status),
    _ => {}
}
```

## frame – where the zone goes

```rust
let (sx, sy) = z.self_pos().unwrap_or((z.width / 2, z.height / 2));
let (x0, y0) = (camera(sx, cols, z.width), camera(sy, rows, z.height));
let pad = (cols.saturating_sub(z.width) / 2, rows.saturating_sub(z.height) / 2);
let grid = render_zone(z, x0, y0, cols - pad.0, rows - pad.1);
```

Grid cell `(x, y)` is zone cell `(x0 + x, y0 + y)` and goes to byte offset
`((y + pad.1) * cols + x + pad.0) * CELL_BYTES`.

## frame – the bytes of a cell

```rust
let b = &mut self.frame[o..o + CELL_BYTES];
b[0..4].copy_from_slice(&(cell.glyph as u32).to_le_bytes());
let fg = shade(cell.fg, light);
b[4..7].copy_from_slice(&[fg.0, fg.1, fg.2]);
```

and the same for `bg` at 7..10, with `b[10] = 1`. `zone_light` only for
cells inside the zone.

## click

Ignore clicks left of or above the zone (`col < pad.0`). Then
`(origin.0 + col - pad.0, origin.1 + row - pad.1)` is the zone cell – walk
there exactly like `App::walk_to` in lesson 30.

## tick

```rust
if self.walk.is_empty() || t_ms - self.last_step < WALK_MS {
    return;
}
```

## "can't find crate for `core`" / "target may not be installed"

`rustup target add wasm32-unknown-unknown`.

## Looking inside the browser

F12 → Console: `wasm.frame(10, 5, 0)` returns a pointer, and
`new Uint8Array(wasm.memory.buffer, ptr, 600)` shows the bytes.
