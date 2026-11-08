//! A *scene* is everything that should be on the screen, described without
//! saying how it is drawn: cells with glyphs, colours and sprite numbers, a
//! status line, log lines, the input line. A `Renderer` draws a scene – as
//! text, in a terminal, or with tiles in a window. The rest of the client
//! doesn't know which.

use crate::app::{shown_input, App, LogLine};
use crate::messages;
use crate::render::{render_zone, Rgb};
use crate::tileset::{entity_sprite, tile_sprite};
use crate::zone::camera;

/// How many log lines the scene shows.
pub const LOG_LINES: u16 = 5;

/// No change to a colour.
pub const FULL_LIGHT: Rgb = Rgb(255, 255, 255);

/// What explored-but-not-visible cells are multiplied by: the fog of war.
pub const FOG: Rgb = Rgb(90, 90, 110);

/// One cell of the zone, ready to draw.
#[derive(Debug, Clone, PartialEq)]
pub struct SceneCell {
    pub x: u16,
    pub y: u16,
    /// For text renderers.
    pub glyph: char,
    pub fg: Rgb,
    pub bg: Option<Rgb>,
    /// For tile renderers: the atlas index of the tile and of the entity on it.
    pub tile: Option<u16>,
    pub entity: Option<u16>,
    /// Multiply the colours by this (fog of war, and light in lesson 29).
    pub light: Rgb,
}

/// Everything on the screen.
#[derive(Debug, Clone, PartialEq, Default)]
pub struct Scene {
    pub width: u16,
    pub height: u16,
    /// The zone's cells, with (0, 0) at the top left of the zone view.
    pub cells: Vec<SceneCell>,
    pub status: String,
    pub log: Vec<LogLine>,
    pub input: String,
}

/// Draws scenes. Implemented by `TextRenderer` (below), and by the window in
/// `gfx.rs` – one trait, several ways to draw.
pub trait Renderer {
    /// Draws one cell of the zone at (x, y) in cells.
    fn cell(&mut self, x: u16, y: u16, cell: &SceneCell);
    /// Writes text starting at (x, y) in cells.
    fn text(&mut self, x: u16, y: u16, text: &str, fg: Rgb);
}

/// Multiplies a colour by a light colour (255 = unchanged).
pub fn shade(c: Rgb, light: Rgb) -> Rgb {
    // TODO(lesson 28): every component: c × light / 255 (in u16, to avoid overflow).
    let _ = (c, light);
    todo!("shade")
}

/// The height of the zone view in a scene of `height` rows: everything but
/// the status line, the log and the input line.
pub fn zone_rows(height: u16) -> u16 {
    height.saturating_sub(2 + LOG_LINES)
}

/// Builds the scene for a screen of width × height cells.
pub fn build_scene(app: &App, width: u16, height: u16) -> Scene {
    // TODO(lesson 28): status, the last LOG_LINES log lines and "> " + shown_input;
    // with a zone: one SceneCell per explored cell in the camera's view (see
    // the README), light FULL_LIGHT if visible, else FOG.
    let _ = (app, width, height, camera, render_zone, tile_sprite, entity_sprite, shown_input, messages::NOT_LOGGED_IN);
    todo!("build_scene")
}

/// Draws a scene: the status line at the top, the zone below it, then the
/// log and the input line at the bottom.
pub fn draw_scene(scene: &Scene, r: &mut impl Renderer) {
    // TODO(lesson 28): status on row 0, the cells one row down, the log under
    // the zone, the input on the last row.
    let _ = (scene, r);
    todo!("draw_scene")
}

/// A renderer that draws into a grid of characters – for tests, and for
/// seeing that the trait really works without a window.
pub struct TextRenderer {
    pub grid: Vec<Vec<char>>,
}

impl TextRenderer {
    pub fn new(width: u16, height: u16) -> TextRenderer {
        TextRenderer {
            grid: vec![vec![' '; width as usize]; height as usize],
        }
    }

    /// The grid as lines.
    pub fn lines(&self) -> Vec<String> {
        self.grid.iter().map(|row| row.iter().collect()).collect()
    }

    fn put(&mut self, x: u16, y: u16, ch: char) {
        if let Some(cell) = self.grid.get_mut(y as usize).and_then(|row| row.get_mut(x as usize)) {
            *cell = ch;
        }
    }
}

impl Renderer for TextRenderer {
    fn cell(&mut self, x: u16, y: u16, cell: &SceneCell) {
        // TODO(lesson 28): put the cell's glyph at (x, y).
        let _ = (x, y, cell, Self::put);
        todo!("TextRenderer::cell")
    }

    fn text(&mut self, x: u16, y: u16, text: &str, _fg: Rgb) {
        // TODO(lesson 28): one character per cell, from (x, y) to the right.
        let _ = (x, y, text);
        todo!("TextRenderer::text")
    }
}
