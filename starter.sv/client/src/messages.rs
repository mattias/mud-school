//! Texts shown by the client, in the course language (svenska).
//! `./mudschool lang <sv|en|ja>` replaces this file.

pub const CONNECTED_TO: &str = "Ansluten till";
pub const SERVER_CLOSED: &str = "Servern stängde anslutningen.";
pub const DISCONNECTED: &str = "*** Anslutningen till servern stängdes. Tryck Esc för att avsluta.";
pub const NOT_LOGGED_IN: &str = " Inte inloggad";
pub const NO_MAP: &str = "(ingen karta)";
pub const TITLE_LOG: &str = " Logg ";
pub const TITLE_MAP: &str = " Karta ";
pub const TITLE_INPUT: &str = " Kommando ";
pub const ONLINE: &str = "online";
pub const STARTER: &str = "Hej från MUD-klienten! Den här filen fyller du på i lektion 3.";
pub const TITLE_PASSWORD: &str = " Lösenord (dolt) ";
pub const CONFIG_ERROR: &str = "Fel i inställningarna";
// The mouse (lesson 30): menu entries and what tiles are called.
pub const LOOK_AT: &str = "Titta på";
pub const ATTACK: &str = "Anfall";
pub const PICK_UP: &str = "Plocka upp";
pub const WALK_HERE: &str = "Gå hit";

/// What a tile meaning is called, for the mouse pointer's tooltip.
pub fn tile_name(meaning: &str) -> &'static str {
    match meaning {
        "floor" => "Golv",
        "wall" => "En vägg",
        "grass" => "Gräs",
        "tree" => "Ett träd",
        "water" => "Vatten",
        "rock" => "Berg",
        "door" => "En dörr",
        "stairs_up" => "Trappa upp",
        "stairs_down" => "Trappa ner",
        _ => "Något",
    }
}
