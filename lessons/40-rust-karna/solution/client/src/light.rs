//! Light and life (lesson 29): particles and a shaking screen. The light
//! itself – sources, ambient light, shimmering water – lives in the core
//! since lesson 40 (`mud_core::light`) and is re-exported here.

use std::time::Instant;

pub use mud_core::light::*;

use crate::render::Rgb;

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
