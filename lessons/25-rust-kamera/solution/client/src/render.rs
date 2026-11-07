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

const fn cell(glyph: char, fg: Rgb) -> Cell {
    Cell {
        glyph,
        fg,
        bg: None,
    }
}

/// How a thing with a certain meaning should look.
pub fn tile(meaning: &str) -> Cell {
    match meaning {
        "town" => cell('#', Rgb(176, 140, 96)),
        "inside" => cell('■', Rgb(214, 180, 130)),
        "road" => cell('·', Rgb(160, 140, 110)),
        "field" => cell('"', Rgb(170, 210, 90)),
        "forest" => cell('♣', Rgb(40, 170, 70)),
        "water" => Cell {
            glyph: '≈',
            fg: Rgb(90, 160, 255),
            bg: Some(Rgb(10, 30, 90)),
        },
        "mountain" => cell('▲', Rgb(175, 175, 190)),
        "cave" => cell('Ω', Rgb(150, 110, 190)),
        "stairs_up" => cell('<', Rgb(240, 240, 240)),
        "stairs_down" => cell('>', Rgb(240, 240, 240)),
        "path_ew" => cell('─', Rgb(110, 100, 90)),
        "path_ns" => cell('│', Rgb(110, 100, 90)),
        // zone tiles (lesson 25)
        "floor" => cell('.', Rgb(120, 110, 100)),
        "wall" => cell('#', Rgb(150, 140, 130)),
        "grass" => cell('"', Rgb(90, 170, 70)),
        "tree" => cell('♣', Rgb(40, 150, 60)),
        "rock" => cell('▲', Rgb(150, 150, 165)),
        "door" => cell('+', Rgb(190, 140, 80)),
        "self" => cell('@', Rgb(255, 230, 50)),
        "player" => cell('&', Rgb(80, 220, 255)),
        _ => cell('?', Rgb(255, 0, 255)),
    }
}

/// How an entity in a zone should look: its own glyph, coloured by kind.
pub fn entity_cell(kind: &str, glyph: &str, is_self: bool) -> Cell {
    let g = glyph.chars().next().unwrap_or('?');
    let fg = match kind {
        _ if is_self => Rgb(255, 230, 50),
        "player" => Rgb(80, 220, 255),
        "creature" => Rgb(230, 90, 80),
        "item" => Rgb(240, 200, 120),
        _ => Rgb(255, 0, 255),
    };
    cell(g, fg)
}

/// The part of a zone that starts at (x0, y0) and is w×h cells, with its
/// entities on top.
pub fn render_zone(z: &crate::zone::ZoneView, x0: usize, y0: usize, w: usize, h: usize) -> Vec<Vec<Cell>> {
    (y0..y0 + h)
        .map(|y| {
            (x0..x0 + w)
                .map(|x| {
                    if let Some(e) = z.entity_at(x, y) {
                        return entity_cell(&e.kind, &e.glyph, e.id == z.self_id);
                    }
                    match z.code_at(x, y) {
                        ' ' => BLANK,
                        code => match z.meaning_at(x, y) {
                            Some(m) => tile(m),
                            None => cell(code, Rgb(128, 128, 128)),
                        },
                    }
                })
                .collect()
        })
        .collect()
}

/// Turns the whole map into a grid of cells.
pub fn render_map(map: &MapData) -> Vec<Vec<Cell>> {
    map.rows
        .iter()
        .map(|row| {
            row.chars()
                .map(|code| match code {
                    ' ' => BLANK,
                    _ => match map.meaning(code) {
                        Some(meaning) => tile(meaning),
                        None => cell(code, Rgb(128, 128, 128)), // unknown code: show it as it is
                    },
                })
                .collect()
        })
        .collect()
}

/// Draws the grid as text with ANSI colour codes (24-bit "truecolor").
/// Rows are separated by '\n'.
pub fn to_ansi(grid: &[Vec<Cell>]) -> String {
    let mut out = String::new();
    for (i, row) in grid.iter().enumerate() {
        if i > 0 {
            out.push('\n');
        }
        for c in row {
            let Rgb(r, g, b) = c.fg;
            write!(out, "\x1b[38;2;{r};{g};{b}m").unwrap();
            if let Some(Rgb(r, g, b)) = c.bg {
                write!(out, "\x1b[48;2;{r};{g};{b}m").unwrap();
            }
            out.push(c.glyph);
            out.push_str("\x1b[0m");
        }
    }
    out
}
