//! The mouse (lesson 30): click to walk (A* in the client), a tooltip that
//! names what is under the pointer, and a right-click menu. The paths
//! themselves live in the core since lesson 40 (`mud_core::path`).

pub use mud_core::path::{find_path, step_command, walkable, Cell, STEPS};

use crate::messages;
use crate::zone::ZoneView;

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
