# Hints

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

`chain` glues two iterators together – added and moved are handled the same way.

## camera

```rust
if size <= view {
    return 0;
}
focus.saturating_sub(view / 2).min(size - view)
```

## The layout

```rust
let [zone_area, log_area] =
    Layout::vertical([Constraint::Min(3), Constraint::Length(log_height)]).areas(left_area);
```

Inside a `Block` with borders, `block.inner(area)` is the space left for the
content – use its width and height as the camera's view.
