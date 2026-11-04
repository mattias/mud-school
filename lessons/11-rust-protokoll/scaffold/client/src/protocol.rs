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
    Text {
        #[serde(default)]
        text: String,
    },
    // TODO(lesson 11): Hello { text }, Map { map: MapData }, Status { status: Status }
    // and an Unknown variant for types we don't know (hint: #[serde(other)]).
}

/// A piece of the map. Every character in `rows` is a code that `legend` explains.
#[derive(Debug, Clone, PartialEq, Default, Deserialize)]
pub struct MapData {
    // TODO(lesson 11): width, height, rows, legend (see docs/<lang>/PROTOCOL.md)
}

impl MapData {
    /// What does the code `code` mean? E.g. 'T' -> Some("forest").
    pub fn meaning(&self, code: char) -> Option<&str> {
        todo!()
    }
}

/// What is shown in the status bar.
#[derive(Debug, Clone, PartialEq, Default, Deserialize)]
pub struct Status {
    // TODO(lesson 11): name, room, terrain, hp, max_hp, online
}

impl fmt::Display for Status {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        // TODO(lesson 11): e.g. "Anna · Square · HP 10/10 · 2 online" (see messages::ONLINE)
        todo!()
    }
}

/// Parses a line from the server. Lines that are not valid JSON (e.g. before
/// the handshake) become plain text – so the client works with any server.
pub fn parse_line(line: &str) -> ServerMsg {
    todo!()
}
