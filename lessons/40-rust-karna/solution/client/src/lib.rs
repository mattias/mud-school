//! The library part of the MUD client. Everything the tests need to reach
//! lives here (and in modules you add with `pub mod ...;`).

// The core (lesson 40), re-exported so that `crate::protocol` and
// `mud_client::zone` mean the same as before.
pub use mud_core::{protocol, render, tileset, zone};

pub mod app;
pub mod config;
pub mod light;
pub mod messages;
pub mod mouse;
pub mod net;
pub mod reader;
pub mod scene;

/// The window (lesson 28) – only with `--features gfx`.
#[cfg(feature = "gfx")]
pub mod gfx;
