//! Protokollet mellan server och klient. Se docs/PROTOKOLL.md.

use std::collections::HashMap;
use std::fmt;

use serde::Deserialize;

/// Det klienten skickar för att få JSON istället för text.
pub const HELLO: &str = "@hello mudschool 1";

/// Ett meddelande från servern.
///
/// `#[serde(tag = "type")]` betyder att JSON-fältet "type" avgör vilken
/// variant det är: `{"type":"text","text":"hej"}` blir `ServerMsg::Text { text: "hej" }`.
#[derive(Debug, Clone, PartialEq, Deserialize)]
#[serde(tag = "type", rename_all = "lowercase")]
pub enum ServerMsg {
    Hello {
        #[serde(default)]
        text: String,
    },
    Text {
        #[serde(default)]
        text: String,
    },
    Map {
        map: MapData,
    },
    Status {
        status: Status,
    },
    /// En typ vi inte känner till än – från en nyare server.
    #[serde(other)]
    Unknown,
}

/// En bit av kartan. Varje tecken i `rows` är en kod som `legend` förklarar.
#[derive(Debug, Clone, PartialEq, Default, Deserialize)]
pub struct MapData {
    pub width: usize,
    pub height: usize,
    pub rows: Vec<String>,
    #[serde(default)]
    pub legend: HashMap<String, String>,
}

impl MapData {
    /// Vad betyder koden `code`? T.ex. 'T' -> Some("forest").
    pub fn meaning(&self, code: char) -> Option<&str> {
        self.legend
            .get(code.to_string().as_str())
            .map(String::as_str)
    }
}

/// Det som visas i statusraden.
#[derive(Debug, Clone, PartialEq, Default, Deserialize)]
pub struct Status {
    pub name: String,
    pub room: String,
    #[serde(default)]
    pub terrain: String,
    pub hp: i32,
    pub max_hp: i32,
    #[serde(default)]
    pub online: u32,
}

impl fmt::Display for Status {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "{} · {} · HP {}/{} · {} online",
            self.name, self.room, self.hp, self.max_hp, self.online
        )
    }
}

/// Tolkar en rad från servern. Rader som inte är giltig JSON (t.ex. innan
/// handskakningen) blir vanlig text – så klienten fungerar med alla servrar.
pub fn parse_line(line: &str) -> ServerMsg {
    if line.starts_with('{') {
        if let Ok(msg) = serde_json::from_str(line) {
            return msg;
        }
    }
    ServerMsg::Text {
        text: line.to_string(),
    }
}
