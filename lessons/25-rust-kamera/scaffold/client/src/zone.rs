//! The zone the player is in, as the client knows it: tiles from the last
//! `zone` frame, kept up to date by `delta` frames. And the camera that
//! decides which part of a big zone fits on the screen.

use std::collections::{BTreeMap, HashMap};

use crate::protocol::{DeltaData, Entity, ZoneData};

/// What the client knows about the current zone.
#[derive(Debug, Clone, Default, PartialEq)]
pub struct ZoneView {
    pub room: String,
    pub width: usize,
    pub height: usize,
    /// rows[y][x] – the tile codes.
    pub rows: Vec<Vec<char>>,
    pub legend: HashMap<String, String>,
    /// Everything with a position, by ID. A BTreeMap keeps them sorted, so
    /// drawing is the same every time.
    pub entities: BTreeMap<String, Entity>,
    /// The player's own entity ID.
    pub self_id: String,
}

impl From<ZoneData> for ZoneView {
    fn from(z: ZoneData) -> ZoneView {
        ZoneView {
            room: z.room,
            width: z.width,
            height: z.height,
            rows: z.rows.iter().map(|r| r.chars().collect()).collect(),
            legend: z.legend,
            entities: z.entities.into_iter().map(|e| (e.id.clone(), e)).collect(),
            self_id: z.self_id,
        }
    }
}

impl ZoneView {
    /// Applies a delta. A delta for another room is ignored (it may arrive
    /// just before the new zone); returns whether it was applied.
    pub fn apply(&mut self, d: DeltaData) -> bool {
        // TODO(lesson 25): ignore a delta for another room (return false);
        // insert added and moved, remove removed.
        let _ = d;
        todo!("ZoneView::apply")
    }

    /// Where the player stands, if known.
    pub fn self_pos(&self) -> Option<(usize, usize)> {
        // TODO(lesson 25): the position of the entity with the ID self_id.
        todo!("ZoneView::self_pos")
    }

    /// The tile code at (x, y); ' ' outside the zone.
    pub fn code_at(&self, x: usize, y: usize) -> char {
        self.rows
            .get(y)
            .and_then(|row| row.get(x))
            .copied()
            .unwrap_or(' ')
    }

    /// What the tile at (x, y) means ("wall", "grass", ...), if known.
    pub fn meaning_at(&self, x: usize, y: usize) -> Option<&str> {
        self.legend
            .get(self.code_at(x, y).to_string().as_str())
            .map(String::as_str)
    }

    /// The entity drawn at (x, y): players over creatures over items.
    pub fn entity_at(&self, x: usize, y: usize) -> Option<&Entity> {
        let rank = |e: &Entity| match e.kind.as_str() {
            "player" => 2,
            "creature" => 1,
            _ => 0,
        };
        self.entities
            .values()
            .filter(|e| e.x as usize == x && e.y as usize == y && e.x >= 0 && e.y >= 0)
            .max_by_key(|e| rank(e))
    }
}

/// The camera along one axis: the first cell to show so that `focus` is in
/// the middle of a view of `view` cells, without showing anything outside a
/// zone of `size` cells. A zone smaller than the view starts at 0.
pub fn camera(focus: usize, view: usize, size: usize) -> usize {
    // TODO(lesson 25): see the comment above and the README.
    let _ = (focus, view, size);
    todo!("camera")
}
