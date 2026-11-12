//! Light (lesson 29): light sources that blend their colours and shimmering
//! water – plain functions of position and time in milliseconds. In the core
//! since lesson 40, with `zone_light`, the light of a cell in a zone.

use crate::render::Rgb;
use crate::zone::ZoneView;

/// A light source: a warm torch, a cold glow, ...
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct Light {
    pub x: f32,
    pub y: f32,
    /// How far it reaches, in cells.
    pub radius: f32,
    pub color: Rgb,
}

/// The torch every player carries.
pub const TORCH: Rgb = Rgb(255, 190, 110);

/// How light it is everywhere, without any light sources, for a terrain.
pub fn ambient(terrain: &str) -> Rgb {
    match terrain {
        "cave" => Rgb(30, 30, 45),
        "mountain" => Rgb(150, 150, 165),
        "forest" => Rgb(120, 140, 120),
        "inside" => Rgb(110, 100, 90),
        "" => Rgb(255, 255, 255), // unknown: full light
        _ => Rgb(220, 220, 210),  // town, road, field, water
    }
}

/// The light at the middle of cell (x, y): the ambient light plus every
/// source's colour, weaker with distance (linearly to 0 at its radius).
/// The channels are added and capped at 255 – a red and a green light make
/// yellow.
pub fn light_at(x: usize, y: usize, lights: &[Light], ambient: Rgb) -> Rgb {
    let (mut r, mut g, mut b) = (ambient.0 as f32, ambient.1 as f32, ambient.2 as f32);
    for l in lights {
        let (dx, dy) = (x as f32 - l.x, y as f32 - l.y);
        let d = (dx * dx + dy * dy).sqrt();
        let k = (1.0 - d / l.radius).max(0.0);
        r += l.color.0 as f32 * k;
        g += l.color.1 as f32 * k;
        b += l.color.2 as f32 * k;
    }
    Rgb(r.min(255.0) as u8, g.min(255.0) as u8, b.min(255.0) as u8)
}

/// The light sources in a zone: every player carries a torch.
pub fn lights_for(zone: &ZoneView) -> Vec<Light> {
    zone.entities
        .values()
        .filter(|e| e.kind == "player")
        .map(|e| Light {
            x: e.x as f32,
            y: e.y as f32,
            radius: 6.0,
            color: TORCH,
        })
        .collect()
}

/// How bright water at (x, y) is at time t (in milliseconds): a slow wave
/// that moves across the lake. 1.0 = normal.
pub fn shimmer(x: usize, y: usize, t_ms: u64) -> f32 {
    let phase = t_ms as f32 / 400.0 + x as f32 * 0.7 + y as f32 * 0.3;
    0.85 + 0.15 * phase.sin()
}

/// Scales a colour (e.g. light) by a factor, capped at 255.
pub fn scale(c: Rgb, k: f32) -> Rgb {
    let s = |v: u8| (v as f32 * k).clamp(0.0, 255.0) as u8;
    Rgb(s(c.0), s(c.1), s(c.2))
}

/// No change to a colour.
pub const FULL_LIGHT: Rgb = Rgb(255, 255, 255);

/// What explored-but-not-visible cells are multiplied by: the fog of war.
pub const FOG: Rgb = Rgb(90, 90, 110);

/// Multiplies a colour by a light colour (255 = unchanged).
pub fn shade(c: Rgb, light: Rgb) -> Rgb {
    let m = |a: u8, b: u8| ((a as u16 * b as u16) / 255) as u8;
    Rgb(m(c.0, light.0), m(c.1, light.1), m(c.2, light.2))
}

/// The light on cell (x, y) of a zone in a room with `terrain`, `t_ms`
/// milliseconds after the client started (lesson 40): fog if the cell isn't
/// visible, otherwise the ambient light plus the torches – and water shimmers.
pub fn zone_light(zone: &ZoneView, terrain: &str, x: usize, y: usize, t_ms: u64) -> Rgb {
    if !zone.is_visible(x, y) {
        return FOG;
    }
    let light = light_at(x, y, &lights_for(zone), ambient(terrain));
    if zone.meaning_at(x, y) == Some("water") {
        return scale(light, shimmer(x, y, t_ms));
    }
    light
}
