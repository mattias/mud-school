# Hints

## find_path

```rust
let h = |c: Cell| (c.0 as i32 - to.0 as i32).abs().max((c.1 as i32 - to.1 as i32).abs()) as u32;
let mut open = BinaryHeap::new();
let mut came: HashMap<Cell, Cell> = HashMap::new();
let mut best: HashMap<Cell, u32> = HashMap::from([(from, 0)]);
open.push(Reverse((h(from), 0u32, from)));
while let Some(Reverse((_, g, cur))) = open.pop() {
    if cur == to {
        // walk back through `came` until `from`, then reverse
    }
    if g > best[&cur] {
        continue;
    }
    for (dx, dy, _) in STEPS {
        // the neighbour, if it is inside (no negative coordinates) and walkable,
        // and if g + 1 is better than what we had: remember it and push it
    }
}
None
```

## The ID without its prefix

```rust
let word = e.id.split_once(':').map_or(e.id.as_str(), |(_, w)| w);
```

## next_walk_step

```rust
if self.walk.is_empty() || self.next_step.is_some_and(|t| now < t) {
    return None;
}
self.next_step = Some(now + WALK_STEP);
self.walk.pop_front()
```

## left_click with a menu

`self.menu.take()` takes the menu out (leaving `None`) – the menu closes
whatever you choose. `?` on an `Option` returns `None` early if there was no
such row.
