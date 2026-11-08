# ヒント

## light_at

```rust
let (mut r, mut g, mut b) = (ambient.0 as f32, ambient.1 as f32, ambient.2 as f32);
for l in lights {
    let (dx, dy) = (x as f32 - l.x, y as f32 - l.y);
    let k = (1.0 - (dx * dx + dy * dy).sqrt() / l.radius).max(0.0);
    r += l.color.0 as f32 * k;
    // g, b
}
Rgb(r.min(255.0) as u8, g.min(255.0) as u8, b.min(255.0) as u8)
```

`f32` で計算し、最後に変換します。`u8` では 200 + 100 がオーバーフローします。

## Particles::update

```rust
for p in &mut self.list {
    p.x += p.vx * dt;
    p.y += p.vy * dt;
    p.vy += 6.0 * dt;
    p.life -= dt;
}
self.list.retain(|p| p.life > 0.0);
```

## shake_offset

```rust
let Some(until) = until else { return (0.0, 0.0) };
if now >= until {
    return (0.0, 0.0);
}
let left = (until - now).as_secs_f32();
let k = (left / 0.3).min(1.0) * 0.25;
let t = left * 60.0;
(t.sin() * k, (t * 1.3).cos() * k)
```

## cell_light の時刻

```rust
let t = app.started.map_or(0, |s| now.saturating_duration_since(s).as_millis() as u64);
```

`saturating_duration_since` は、`now` のほうが早くても panic せず 0 を返します。
