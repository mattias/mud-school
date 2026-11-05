//! The protocol between server and client. See docs/<lang>/PROTOCOL.md.

use std::collections::HashMap;
use std::fmt;

use serde::Deserialize;

/// What the client sends to get JSON instead of text.
pub const HELLO: &str = "@hello mudschool 1";

/// A message from the server.
///
/// `#[serde(tag = "type")]` means that the JSON field "type" decides which
/// variant it is: `{"type":"text","text":"hi"}` becomes `ServerMsg::Text { text: "hi" }`.
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
        /// What kind of text it is (lesson 17). Old servers don't send it.
        #[serde(default)]
        kind: Kind,
    },
    Map {
        map: MapData,
    },
    Status {
        status: Status,
    },
    /// A type we don't know yet – from a newer server.
    #[serde(other)]
    Unknown,
}

/// What kind of text a text frame is, so that the client can colour it.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Default, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum Kind {
    /// Everything else – also what a missing `kind` means.
    #[default]
    Info,
    Room,
    Say,
    Move,
    Error,
    Combat,
    /// A kind from a newer server: show it like Info.
    #[serde(other)]
    Other,
}

/// A piece of the map. Every character in `rows` is a code that `legend` explains.
#[derive(Debug, Clone, PartialEq, Default, Deserialize)]
pub struct MapData {
    pub width: usize,
    pub height: usize,
    pub rows: Vec<String>,
    #[serde(default)]
    pub legend: HashMap<String, String>,
}

impl MapData {
    /// What does the code `code` mean? E.g. 'T' -> Some("forest").
    pub fn meaning(&self, code: char) -> Option<&str> {
        self.legend
            .get(code.to_string().as_str())
            .map(String::as_str)
    }
}

/// What is shown in the status bar.
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
    /// Who you are fighting (lesson 19). Missing = nobody.
    #[serde(default)]
    pub target: Option<Target>,
}

/// The opponent in a fight.
#[derive(Debug, Clone, PartialEq, Eq, Default, Deserialize)]
pub struct Target {
    pub name: String,
    pub hp: i32,
    pub max_hp: i32,
}

impl fmt::Display for Status {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "{} · {} · HP {}/{} · {} {}",
            self.name,
            self.room,
            self.hp,
            self.max_hp,
            self.online,
            crate::messages::ONLINE
        )
    }
}

/// Parses a line from the server. Lines that are not valid JSON (e.g. before
/// the handshake) become plain text – so the client works with any server.
pub fn parse_line(line: &str) -> ServerMsg {
    if line.starts_with('{') {
        if let Ok(msg) = serde_json::from_str(line) {
            return msg;
        }
    }
    ServerMsg::Text {
        text: line.to_string(),
        kind: Kind::Info,
    }
}
