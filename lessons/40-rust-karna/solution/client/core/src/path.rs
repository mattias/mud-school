//! Paths in a zone (lesson 30): which tiles can be walked on, and A* from
//! one cell to another. In the core since lesson 40, so that the browser
//! client can walk too.

use std::cmp::Reverse;
use std::collections::{BinaryHeap, HashMap};

use crate::zone::ZoneView;

/// A cell in the zone.
pub type Cell = (usize, usize);

/// Whether a tile meaning can be walked on (the same rule as the server's
/// `Walkable`, lesson 23).
pub fn walkable(meaning: Option<&str>) -> bool {
    matches!(meaning, Some("floor" | "grass" | "door" | "stairs_up" | "stairs_down"))
}

/// The eight directions as (dx, dy) and the command that walks that way.
pub const STEPS: [(i32, i32, &str); 8] = [
    (0, -1, "north"),
    (1, -1, "northeast"),
    (1, 0, "east"),
    (1, 1, "southeast"),
    (0, 1, "south"),
    (-1, 1, "southwest"),
    (-1, 0, "west"),
    (-1, -1, "northwest"),
];

/// The command for one step from `from` to the neighbouring cell `to`.
pub fn step_command(from: Cell, to: Cell) -> Option<&'static str> {
    let (dx, dy) = (to.0 as i32 - from.0 as i32, to.1 as i32 - from.1 as i32);
    STEPS.iter().find(|s| s.0 == dx && s.1 == dy).map(|s| s.2)
}

/// The shortest path from `from` to `to` over tiles the client knows are
/// walkable (diagonal steps allowed), without `from`, with `to`. A* with a
/// BinaryHeap – Rust's priority queue is a max-heap, so `Reverse` makes the
/// smallest f come out first.
pub fn find_path(zone: &ZoneView, from: Cell, to: Cell) -> Option<Vec<Cell>> {
    if !walkable(zone.meaning_at(to.0, to.1)) {
        return None;
    }
    if from == to {
        return Some(Vec::new());
    }
    let h = |c: Cell| (c.0 as i32 - to.0 as i32).abs().max((c.1 as i32 - to.1 as i32).abs()) as u32;
    let mut open = BinaryHeap::new();
    let mut came: HashMap<Cell, Cell> = HashMap::new();
    let mut best: HashMap<Cell, u32> = HashMap::from([(from, 0)]);
    open.push(Reverse((h(from), 0u32, from)));
    while let Some(Reverse((_, g, cur))) = open.pop() {
        if cur == to {
            let mut path = vec![to];
            let mut c = to;
            while let Some(&prev) = came.get(&c) {
                if prev == from {
                    break;
                }
                path.push(prev);
                c = prev;
            }
            path.reverse();
            return Some(path);
        }
        if g > best[&cur] {
            continue;
        }
        for (dx, dy, _) in STEPS {
            let (nx, ny) = (cur.0 as i32 + dx, cur.1 as i32 + dy);
            if nx < 0 || ny < 0 {
                continue;
            }
            let next = (nx as usize, ny as usize);
            if !walkable(zone.meaning_at(next.0, next.1)) {
                continue;
            }
            let ng = g + 1;
            if best.get(&next).is_some_and(|&old| ng >= old) {
                continue;
            }
            best.insert(next, ng);
            came.insert(next, cur);
            open.push(Reverse((ng + h(next), ng, next)));
        }
    }
    None
}

