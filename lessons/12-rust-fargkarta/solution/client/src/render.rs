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

const fn cell(glyph: char, fg: Rgb) -> Cell {
    Cell {
        glyph,
        fg,
        bg: None,
    }
}

/// Hur en sak med en viss betydelse ska se ut.
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
        "self" => cell('@', Rgb(255, 230, 50)),
        "player" => cell('&', Rgb(80, 220, 255)),
        _ => cell('?', Rgb(255, 0, 255)),
    }
}

/// Gör om hela kartan till ett rutnät av celler.
pub fn render_map(map: &MapData) -> Vec<Vec<Cell>> {
    map.rows
        .iter()
        .map(|row| {
            row.chars()
                .map(|code| match code {
                    ' ' => BLANK,
                    _ => match map.meaning(code) {
                        Some(meaning) => tile(meaning),
                        None => cell(code, Rgb(128, 128, 128)), // okänd kod: visa den som den är
                    },
                })
                .collect()
        })
        .collect()
}

/// Ritar rutnätet som text med ANSI-färgkoder (24-bitars "truecolor").
/// Raderna skiljs åt med '\n'.
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
