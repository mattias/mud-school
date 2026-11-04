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
    Text {
        #[serde(default)]
        text: String,
    },
    // TODO(lektion 11): Hello { text }, Map { map: MapData }, Status { status: Status }
    // och en Unknown-variant för typer vi inte känner till (tips: #[serde(other)]).
}

/// En bit av kartan. Varje tecken i `rows` är en kod som `legend` förklarar.
#[derive(Debug, Clone, PartialEq, Default, Deserialize)]
pub struct MapData {
    // TODO(lektion 11): width, height, rows, legend (se docs/PROTOKOLL.md)
}

impl MapData {
    /// Vad betyder koden `code`? T.ex. 'T' -> Some("forest").
    pub fn meaning(&self, code: char) -> Option<&str> {
        todo!()
    }
}

/// Det som visas i statusraden.
#[derive(Debug, Clone, PartialEq, Default, Deserialize)]
pub struct Status {
    // TODO(lektion 11): name, room, terrain, hp, max_hp, online
}

impl fmt::Display for Status {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        // TODO(lektion 11): t.ex. "Anna · Torget · HP 10/10 · 2 online"
        todo!()
    }
}

/// Tolkar en rad från servern. Rader som inte är giltig JSON (t.ex. innan
/// handskakningen) blir vanlig text – så klienten fungerar med alla servrar.
pub fn parse_line(line: &str) -> ServerMsg {
    todo!()
}
