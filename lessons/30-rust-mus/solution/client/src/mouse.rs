//! The mouse (lesson 30): click to walk (A* in the client), a tooltip that
//! names what is under the pointer, and a right-click menu.

use std::cmp::Reverse;
use std::collections::{BinaryHeap, HashMap};

use crate::messages;
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

/// What the tooltip says about a cell: the name of what stands there, or of
/// the tile. Nothing for cells never seen.
pub fn describe_cell(zone: &ZoneView, x: usize, y: usize) -> Option<String> {
    if let Some(e) = zone.entity_at(x, y) {
        return Some(e.name.clone());
    }
    zone.meaning_at(x, y).map(|m| messages::tile_name(m).to_string())
}

/// An entry in the right-click menu.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct MenuItem {
    pub label: String,
    /// What to do: a command for the server, or "@walk" (handled by the client).
    pub command: String,
}

/// The command that walks to the clicked cell – handled by the client itself.
pub const WALK: &str = "@walk";

/// The right-click menu for a cell.
pub fn context_menu(zone: &ZoneView, x: usize, y: usize) -> Vec<MenuItem> {
    let item = |label: String, command: String| MenuItem { label, command };
    let mut menu = Vec::new();
    if let Some(e) = zone.entity_at(x, y) {
        // The ID without its prefix is always a word the server understands.
        let word = e.id.split_once(':').map_or(e.id.as_str(), |(_, w)| w);
        match e.kind.as_str() {
            "creature" => {
                menu.push(item(format!("{} {}", messages::LOOK_AT, e.name), format!("examine {word}")));
                menu.push(item(format!("{} {}", messages::ATTACK, e.name), format!("kill {word}")));
            }
            "item" => {
                menu.push(item(format!("{} {}", messages::LOOK_AT, e.name), format!("examine {word}")));
                menu.push(item(format!("{} {}", messages::PICK_UP, e.name), format!("get {word}")));
            }
            _ => {}
        }
    }
    if walkable(zone.meaning_at(x, y)) {
        menu.push(item(messages::WALK_HERE.to_string(), WALK.to_string()));
    }
    menu
}

/// Converts a mouse position in pixels to a cell of the screen grid.
pub fn screen_to_cell(mx: f32, my: f32, cell_px: f32) -> (u16, u16) {
    ((mx.max(0.0) / cell_px) as u16, (my.max(0.0) / cell_px) as u16)
}
