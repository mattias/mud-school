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
    // TODO(lesson 30): find (dx, dy) in STEPS.
    let _ = (from, to);
    todo!("step_command")
}

/// The shortest path from `from` to `to` over tiles the client knows are
/// walkable (diagonal steps allowed), without `from`, with `to`. A* with a
/// BinaryHeap – Rust's priority queue is a max-heap, so `Reverse` makes the
/// smallest f come out first.
pub fn find_path(zone: &ZoneView, from: Cell, to: Cell) -> Option<Vec<Cell>> {
    // TODO(lesson 30): A* like the server's (lesson 27), with a BinaryHeap of
    // Reverse((f, g, cell)), STEPS as neighbours and walkable(meaning_at(...)).
    let _ = (zone, from, to, BinaryHeap::<Reverse<u32>>::new(), HashMap::<Cell, Cell>::new());
    todo!("find_path")
}

/// What the tooltip says about a cell: the name of what stands there, or of
/// the tile. Nothing for cells never seen.
pub fn describe_cell(zone: &ZoneView, x: usize, y: usize) -> Option<String> {
    // TODO(lesson 30): the entity's name, else messages::tile_name(meaning), else None.
    let _ = (zone, x, y, messages::tile_name);
    todo!("describe_cell")
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
    // TODO(lesson 30): see the README – creature: look + attack, item: look +
    // pick up, then WALK if the cell is walkable.
    let _ = (zone, x, y);
    todo!("context_menu")
}

/// Converts a mouse position in pixels to a cell of the screen grid.
pub fn screen_to_cell(mx: f32, my: f32, cell_px: f32) -> (u16, u16) {
    ((mx.max(0.0) / cell_px) as u16, (my.max(0.0) / cell_px) as u16)
}
