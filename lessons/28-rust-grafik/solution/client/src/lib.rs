//! The library part of the MUD client. Everything the tests need to reach
//! lives here (and in modules you add with `pub mod ...;`).

pub mod app;
pub mod config;
pub mod messages;
pub mod net;
pub mod protocol;
pub mod reader;
pub mod render;
pub mod scene;
pub mod tileset;
pub mod zone;

/// The window (lesson 28) – only with `--features gfx`.
#[cfg(feature = "gfx")]
pub mod gfx;
