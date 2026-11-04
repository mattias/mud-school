//! Turns the map's codes into coloured characters ("tiles").
//!
//! The server says *what* is there ("forest"), this module decides how it
//! should *look*. Change glyphs and colours as you like – it's your client!

use std::fmt::Write;

use crate::protocol::MapData;

/// A colour with red, green and blue components (0–255).
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Rgb(pub u8, pub u8, pub u8);

/// A cell on the screen: a character with a foreground colour and maybe a background colour.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Cell {
    pub glyph: char,
    pub fg: Rgb,
    pub bg: Option<Rgb>, // None = the terminal's normal background
}

/// An empty cell.
pub const BLANK: Cell = Cell {
    glyph: ' ',
    fg: Rgb(0, 0, 0),
    bg: None,
};

/// How a thing with a certain meaning should look.
///
/// Meanings: town, inside, road, field, forest, water, mountain, cave,
/// stairs_up, stairs_down, path_ew, path_ns, self, player.
/// Inspiration: '♣' '≈' '▲' '·' '■' 'Ω' '─' '│' (code page 437, like Dwarf Fortress).
pub fn tile(meaning: &str) -> Cell {
    // TODO(lesson 12)
    todo!()
}

/// Turns the whole map into a grid of cells.
pub fn render_map(map: &MapData) -> Vec<Vec<Cell>> {
    // TODO(lesson 12)
    todo!()
}

/// Draws the grid as text with ANSI colour codes (24-bit "truecolor").
/// Rows are separated by '\n'.
pub fn to_ansi(grid: &[Vec<Cell>]) -> String {
    let mut out = String::new();
    // TODO(lesson 12): for every cell: "\x1b[38;2;R;G;Bm", maybe
    // "\x1b[48;2;R;G;Bm", the character, and finally "\x1b[0m".
    // Hint: write!(out, "...") works on a String thanks to fmt::Write.
    out
}
