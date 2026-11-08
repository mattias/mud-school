//! Texts shown by the client, in the course language (English).
//! `./mudschool lang <sv|en|ja>` replaces this file.

pub const CONNECTED_TO: &str = "Connected to";
pub const SERVER_CLOSED: &str = "The server closed the connection.";
pub const DISCONNECTED: &str = "*** The connection to the server was closed. Press Esc to quit.";
pub const NOT_LOGGED_IN: &str = " Not logged in";
pub const NO_MAP: &str = "(no map)";
pub const TITLE_LOG: &str = " Log ";
pub const TITLE_MAP: &str = " Map ";
pub const TITLE_INPUT: &str = " Command ";
pub const ONLINE: &str = "online";
pub const STARTER: &str = "Hello from the MUD client! You will fill in this file in lesson 3.";
pub const TITLE_PASSWORD: &str = " Password (hidden) ";
pub const CONFIG_ERROR: &str = "Error in the settings";
// The mouse (lesson 30): menu entries and what tiles are called.
pub const LOOK_AT: &str = "Look at";
pub const ATTACK: &str = "Attack";
pub const PICK_UP: &str = "Pick up";
pub const WALK_HERE: &str = "Walk here";

/// What a tile meaning is called, for the mouse pointer's tooltip.
pub fn tile_name(meaning: &str) -> &'static str {
    match meaning {
        "floor" => "Floor",
        "wall" => "A wall",
        "grass" => "Grass",
        "tree" => "A tree",
        "water" => "Water",
        "rock" => "Rock",
        "door" => "A door",
        "stairs_up" => "Stairs up",
        "stairs_down" => "Stairs down",
        _ => "Something",
    }
}
