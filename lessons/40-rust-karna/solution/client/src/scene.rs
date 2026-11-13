//! A *scene* is everything that should be on the screen, described without
//! saying how it is drawn: cells with glyphs, colours and sprite numbers, a
//! status line, log lines, the input line. A `Renderer` draws a scene – as
//! text, in a terminal, or with tiles in a window. The rest of the client
//! doesn't know which.

use std::time::Instant;

use crate::app::{shown_input, App, LogLine};
use crate::light::{shake_offset, zone_light};
pub use crate::light::{shade, FOG, FULL_LIGHT};
use crate::mouse::describe_cell;
use crate::messages;
use crate::protocol::Status;
use crate::render::{render_zone, Rgb};
use crate::tileset::{entity_sprite, tile_sprite};
use crate::zone::camera;

/// How many log lines the scene shows.
pub const LOG_LINES: u16 = 5;

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
    /// Sparks: position in view cells, colour, alpha (lesson 29).
    pub particles: Vec<(f32, f32, Rgb, f32)>,
    /// How far (in cells) the zone is shaken right now (lesson 29).
    pub shake: (f32, f32),
    /// The zone cell shown at view (0, 0) – to turn a mouse click into a zone cell (lesson 30).
    pub origin: (usize, usize),
    /// Text on top of everything: the tooltip and the menu (lesson 30).
    pub overlay: Vec<(u16, u16, String)>,
}

impl Scene {
    /// The zone cell at screen cell (x, y), if (x, y) is inside the zone view.
    pub fn zone_cell(&self, x: u16, y: u16) -> Option<(usize, usize)> {
        if y == 0 || y > zone_rows(self.height) {
            return None;
        }
        Some((self.origin.0 + x as usize, self.origin.1 + (y - 1) as usize))
    }
}

/// Draws scenes. Implemented by `TextRenderer` (below), and by the window in
/// `gfx.rs` – one trait, several ways to draw.
pub trait Renderer {
    /// Draws one cell of the zone at (x, y) in cells.
    fn cell(&mut self, x: u16, y: u16, cell: &SceneCell);
    /// Writes text starting at (x, y) in cells.
    fn text(&mut self, x: u16, y: u16, text: &str, fg: Rgb);
    /// Draws a spark at (x, y) in cells (lesson 29). A *default method*:
    /// renderers that can't draw sparks (like text) don't have to.
    fn particle(&mut self, _x: f32, _y: f32, _color: Rgb, _alpha: f32) {}
    /// Draws text on top of everything, e.g. a tooltip (lesson 30). By
    /// default it is ordinary text.
    fn overlay(&mut self, x: u16, y: u16, text: &str) {
        self.text(x, y, text, Rgb(255, 255, 200));
    }
}

/// The height of the zone view in a scene of `height` rows: everything but
/// the status line, the log and the input line.
pub fn zone_rows(height: u16) -> u16 {
    height.saturating_sub(2 + LOG_LINES)
}

/// Builds the scene for a screen of width × height cells.
pub fn build_scene(app: &App, width: u16, height: u16) -> Scene {
    build_scene_at(app, width, height, Instant::now())
}

/// The light on cell (x, y) of the zone at time `now`: fog if it isn't
/// visible, otherwise the ambient light of the terrain plus the torches –
/// and water shimmers.
pub fn cell_light(app: &App, x: usize, y: usize, now: Instant) -> Rgb {
    let Some(zone) = &app.zone else {
        return FULL_LIGHT;
    };
    let terrain = app.status.as_ref().map_or("", |s| s.terrain.as_str());
    let t = app
        .started
        .map_or(0, |s| now.saturating_duration_since(s).as_millis() as u64);
    zone_light(zone, terrain, x, y, t) // the core does the work (lesson 40)
}

/// The status line: name, room, health and how many are online. (Before
/// lesson 40 this was `impl Display for Status` – but Status lives in the
/// core now, and Rust's orphan rule only lets the crate that owns a type or
/// the trait implement one for the other.)
pub fn status_line(s: &Status) -> String {
    format!(
        "{} · {} · HP {}/{} · {} {}",
        s.name,
        s.room,
        s.hp,
        s.max_hp,
        s.online,
        messages::ONLINE
    )
}

/// Like build_scene, with the time passed in (so tests control the clock).
pub fn build_scene_at(app: &App, width: u16, height: u16, now: Instant) -> Scene {
    let mut scene = Scene {
        width,
        height,
        status: match &app.status {
            Some(s) => status_line(s),
            None => messages::NOT_LOGGED_IN.trim().to_string(),
        },
        log: app
            .log
            .iter()
            .rev()
            .take(LOG_LINES as usize)
            .rev()
            .cloned()
            .collect(),
        input: format!("> {}", shown_input(app)),
        ..Scene::default()
    };
    let Some(zone) = &app.zone else {
        return scene;
    };
    let (w, h) = (width as usize, zone_rows(height) as usize);
    let (sx, sy) = zone.self_pos().unwrap_or((0, 0));
    let (x0, y0) = (camera(sx, w, zone.width), camera(sy, h, zone.height));
    let (vw, vh) = (w.min(zone.width), h.min(zone.height));
    let grid = render_zone(zone, x0, y0, vw, vh);
    for (vy, row) in grid.iter().enumerate() {
        for (vx, c) in row.iter().enumerate() {
            let (zx, zy) = (x0 + vx, y0 + vy);
            if zone.code_at(zx, zy) == ' ' {
                continue; // never seen
            }
            scene.cells.push(SceneCell {
                x: vx as u16,
                y: vy as u16,
                glyph: c.glyph,
                fg: c.fg,
                bg: c.bg,
                tile: zone.meaning_at(zx, zy).and_then(tile_sprite),
                entity: zone
                    .entity_at(zx, zy)
                    .map(|e| entity_sprite(&e.kind, e.id == zone.self_id)),
                light: cell_light(app, zx, zy, now),
            });
        }
    }
    scene.particles = app
        .particles
        .list
        .iter()
        .map(|p| (p.x - x0 as f32, p.y - y0 as f32, p.color, p.alpha()))
        .collect();
    scene.shake = shake_offset(app.flash_until, now);
    scene.origin = (x0, y0);
    // The tooltip next to the pointer, or the open menu (lesson 30).
    let to_screen = |(zx, zy): (usize, usize)| ((zx.saturating_sub(x0)) as u16 + 1, (zy.saturating_sub(y0)) as u16 + 1);
    if let Some(menu) = &app.menu {
        let (sx, sy) = to_screen(menu.cell);
        for (i, item) in menu.items.iter().enumerate() {
            scene.overlay.push((sx, sy + i as u16, format!(" {} ", item.label)));
        }
    } else if let Some((hx, hy)) = app.hover {
        if let Some(text) = describe_cell(zone, hx, hy) {
            let (sx, sy) = to_screen((hx, hy));
            scene.overlay.push((sx, sy, format!(" {text} ")));
        }
    }
    scene
}

/// Draws a scene: the status line at the top, the zone below it, then the
/// log and the input line at the bottom.
pub fn draw_scene(scene: &Scene, r: &mut impl Renderer) {
    r.text(0, 0, &scene.status, Rgb(255, 255, 255));
    for c in &scene.cells {
        r.cell(c.x, c.y + 1, c);
    }
    for &(x, y, color, alpha) in &scene.particles {
        r.particle(x, y + 1.0, color, alpha);
    }
    let log_top = 1 + zone_rows(scene.height);
    for (i, line) in scene.log.iter().enumerate() {
        r.text(0, log_top + i as u16, &line.text, Rgb(200, 200, 200));
    }
    r.text(0, scene.height.saturating_sub(1), &scene.input, Rgb(255, 255, 255));
    for (x, y, text) in &scene.overlay {
        r.overlay(*x, *y, text);
    }
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
        self.put(x, y, cell.glyph);
    }

    fn text(&mut self, x: u16, y: u16, text: &str, _fg: Rgb) {
        for (i, ch) in text.chars().enumerate() {
            self.put(x + i as u16, y, ch);
        }
    }
}
