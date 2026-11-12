//! The client's core (lesson 40): the protocol, the zone, how things look,
//! light and paths. Plain data and functions – no terminal, no window, no
//! network and no clock (time comes in as milliseconds). That is what lets
//! the same code run in the terminal client, the window and the browser.

pub mod light;
pub mod path;
pub mod protocol;
pub mod render;
pub mod tileset;
pub mod zone;
