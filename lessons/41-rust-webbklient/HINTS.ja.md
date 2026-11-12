# ヒント

## on_line

```rust
match parse_line(line) {
    ServerMsg::Zone { zone } => {
        if self.zone.as_ref().is_some_and(|z| z.room != zone.room) {
            self.walk.clear();
        }
        self.zone = Some(ZoneView::from(zone));
    }
    ServerMsg::Delta { delta } => { /* 適用する */ }
    ServerMsg::Status { status } => self.status = Some(status),
    _ => {}
}
```

## frame – ゾーンの置き場所

```rust
let (sx, sy) = z.self_pos().unwrap_or((z.width / 2, z.height / 2));
let (x0, y0) = (camera(sx, cols, z.width), camera(sy, rows, z.height));
let pad = (cols.saturating_sub(z.width) / 2, rows.saturating_sub(z.height) / 2);
let grid = render_zone(z, x0, y0, cols - pad.0, rows - pad.1);
```

格子のマス `(x, y)` はゾーンのマス `(x0 + x, y0 + y)` で、バイト位置
`((y + pad.1) * cols + x + pad.0) * CELL_BYTES` に入ります。

## frame – 1 マスのバイト

```rust
let b = &mut self.frame[o..o + CELL_BYTES];
b[0..4].copy_from_slice(&(cell.glyph as u32).to_le_bytes());
let fg = shade(cell.fg, light);
b[4..7].copy_from_slice(&[fg.0, fg.1, fg.2]);
```

`bg` も同じように 7..10 に入れ、`b[10] = 1`。`zone_light` はゾーンの中のマスだけ。

## click

ゾーンの左や上のクリック（`col < pad.0`）は無視します。それなら
`(origin.0 + col - pad.0, origin.1 + row - pad.1)` がゾーンのマスです。
レッスン 30 の `App::walk_to` とまったく同じようにそこへ歩きましょう。

## tick

```rust
if self.walk.is_empty() || t_ms - self.last_step < WALK_MS {
    return;
}
```

## "can't find crate for `core`" / "target may not be installed"

`rustup target add wasm32-unknown-unknown`。

## ブラウザの中をのぞく

F12 → Console: `wasm.frame(10, 5, 0)` はポインターを返し、
`new Uint8Array(wasm.memory.buffer, ptr, 600)` でバイトが見られます。
