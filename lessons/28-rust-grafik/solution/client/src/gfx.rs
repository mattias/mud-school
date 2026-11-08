//! The window renderer (lesson 28): draws a `Scene` with tiles from the atlas
//! using macroquad. Only compiled with `--features gfx`.

use macroquad::prelude::*;

use crate::render::Rgb;
use crate::scene::{shade, Renderer, SceneCell};
use crate::tileset::{atlas_rect, TILE_PX};

/// Draws scenes in a macroquad window.
pub struct GfxRenderer<'a> {
    pub atlas: &'a Texture2D,
    /// How many screen pixels one atlas pixel becomes (2 = 32×32 cells).
    pub scale: f32,
}

impl GfxRenderer<'_> {
    /// The size of a cell on the screen, in pixels.
    pub fn cell_px(&self) -> f32 {
        TILE_PX * self.scale
    }
}

fn color(c: Rgb) -> Color {
    Color::from_rgba(c.0, c.1, c.2, 255)
}

impl Renderer for GfxRenderer<'_> {
    fn cell(&mut self, x: u16, y: u16, cell: &SceneCell) {
        let size = self.cell_px();
        let (px, py) = (x as f32 * size, y as f32 * size);
        let tint = color(cell.light);
        let sprite = |index: u16| {
            let (sx, sy, sw, sh) = atlas_rect(index);
            draw_texture_ex(
                self.atlas,
                px,
                py,
                tint,
                DrawTextureParams {
                    source: Some(Rect::new(sx, sy, sw, sh)),
                    dest_size: Some(vec2(size, size)),
                    ..Default::default()
                },
            );
        };
        match cell.tile {
            Some(index) => sprite(index),
            None => {
                // No picture for this tile: a coloured glyph, like in the terminal.
                if let Some(bg) = cell.bg {
                    draw_rectangle(px, py, size, size, color(shade(bg, cell.light)));
                }
                draw_text(cell.glyph.encode_utf8(&mut [0; 4]), px + size * 0.25, py + size * 0.8, size, color(shade(cell.fg, cell.light)));
            }
        }
        if let Some(index) = cell.entity {
            sprite(index);
        }
    }

    fn text(&mut self, x: u16, y: u16, text: &str, fg: Rgb) {
        let size = self.cell_px();
        draw_text(text, x as f32 * size + 4.0, y as f32 * size + size * 0.75, size * 0.8, color(fg));
    }
}
