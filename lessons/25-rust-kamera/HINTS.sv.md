# Ledtrådar

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

`chain` klistrar ihop två iteratorer – added och moved hanteras på samma sätt.

## camera

```rust
if size <= view {
    return 0;
}
focus.saturating_sub(view / 2).min(size - view)
```

## Layouten

```rust
let [zone_area, log_area] =
    Layout::vertical([Constraint::Min(3), Constraint::Length(log_height)]).areas(left_area);
```

Inne i ett `Block` med kanter är `block.inner(area)` utrymmet som blir kvar
för innehållet – använd dess bredd och höjd som kamerans vy.
