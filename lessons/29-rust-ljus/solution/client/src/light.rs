//! Light and life (lesson 29): light sources that blend their colours,
//! shimmering water, particles and a shaking screen. All of it is plain
//! functions of position and *time* – so it can be tested without a window.

use std::time::Instant;

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

/// A spark that flies away and fades.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct Particle {
    pub x: f32,
    pub y: f32,
    pub vx: f32,
    pub vy: f32,
    /// Seconds left to live, and how long it lived from the start.
    pub life: f32,
    pub max_life: f32,
    pub color: Rgb,
}

impl Particle {
    /// 1.0 when it is born, 0.0 when it dies – for fading out.
    pub fn alpha(&self) -> f32 {
        (self.life / self.max_life).clamp(0.0, 1.0)
    }
}

/// All the particles that are alive.
#[derive(Debug, Clone, Default, PartialEq)]
pub struct Particles {
    pub list: Vec<Particle>,
}

impl Particles {
    /// Adds n sparks at (x, y) flying out in a circle.
    pub fn burst(&mut self, x: f32, y: f32, n: usize, color: Rgb) {
        for i in 0..n {
            let angle = i as f32 / n as f32 * std::f32::consts::TAU;
            let speed = 2.0 + (i % 3) as f32;
            self.list.push(Particle {
                x,
                y,
                vx: angle.cos() * speed,
                vy: angle.sin() * speed,
                life: 0.6,
                max_life: 0.6,
                color,
            });
        }
    }

    /// Moves everything dt seconds forward; dead particles disappear.
    pub fn update(&mut self, dt: f32) {
        for p in &mut self.list {
            p.x += p.vx * dt;
            p.y += p.vy * dt;
            p.vy += 6.0 * dt; // a little gravity
            p.life -= dt;
        }
        self.list.retain(|p| p.life > 0.0);
    }
}

/// How far (in cells) the screen is shaken at time `now` while it flashes
/// until `until`: back and forth, smaller and smaller.
pub fn shake_offset(until: Option<Instant>, now: Instant) -> (f32, f32) {
    let Some(until) = until else {
        return (0.0, 0.0);
    };
    if now >= until {
        return (0.0, 0.0);
    }
    let left = (until - now).as_secs_f32();
    let k = (left / 0.3).min(1.0) * 0.25;
    let t = left * 60.0;
    (t.sin() * k, (t * 1.3).cos() * k)
}
