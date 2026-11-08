# ヒント

## apply

```rust
if d.room != self.room {
    return false;
}
for e in d.added.into_iter().chain(d.moved) {
    self.entities.insert(e.id.clone(), e);
}
for id in &d.removed {
    self.entities.remove(id);
}
true
```

`chain` は 2 つのイテレーターをつなげます。added と moved は同じように扱います。

## camera

```rust
if size <= view {
    return 0;
}
focus.saturating_sub(view / 2).min(size - view)
```

## レイアウト

```rust
let [zone_area, log_area] =
    Layout::vertical([Constraint::Min(3), Constraint::Length(log_height)]).areas(left_area);
```

枠付きの `Block` の中では、`block.inner(area)` が中身に使える領域です。その幅と高さを
カメラの視野に使います。
