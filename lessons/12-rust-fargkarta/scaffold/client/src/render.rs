//! Gör om kartans koder till färgade tecken ("tiles").
//!
//! Servern säger *vad* som finns ("forest"), den här modulen bestämmer hur
//! det ska *se ut*. Byt glyfer och färger hur du vill – det är din klient!

use std::fmt::Write;

use crate::protocol::MapData;

/// En färg med röd, grön och blå komponent (0–255).
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Rgb(pub u8, pub u8, pub u8);

/// En ruta på skärmen: ett tecken med förgrundsfärg och kanske bakgrundsfärg.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Cell {
    pub glyph: char,
    pub fg: Rgb,
    pub bg: Option<Rgb>, // None = terminalens vanliga bakgrund
}

/// En tom ruta.
pub const BLANK: Cell = Cell {
    glyph: ' ',
    fg: Rgb(0, 0, 0),
    bg: None,
};

/// Hur en sak med en viss betydelse ska se ut.
///
/// Betydelser: town, inside, road, field, forest, water, mountain, cave,
/// stairs_up, stairs_down, path_ew, path_ns, self, player.
/// Inspiration: '♣' '≈' '▲' '·' '■' 'Ω' '─' '│' (kodsida 437, som Dwarf Fortress).
pub fn tile(meaning: &str) -> Cell {
    // TODO(lektion 12)
    todo!()
}

/// Gör om hela kartan till ett rutnät av celler.
pub fn render_map(map: &MapData) -> Vec<Vec<Cell>> {
    // TODO(lektion 12)
    todo!()
}

/// Ritar rutnätet som text med ANSI-färgkoder (24-bitars "truecolor").
/// Raderna skiljs åt med '\n'.
pub fn to_ansi(grid: &[Vec<Cell>]) -> String {
    let mut out = String::new();
    // TODO(lektion 12): för varje cell: "\x1b[38;2;R;G;Bm", kanske
    // "\x1b[48;2;R;G;Bm", tecknet, och sist "\x1b[0m".
    // Tips: write!(out, "...")  fungerar på en String tack vare fmt::Write.
    out
}
