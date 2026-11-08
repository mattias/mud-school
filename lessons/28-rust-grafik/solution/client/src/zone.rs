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
    /// visible[y][x]: what the player sees right now. Empty = everything
    /// (a server from before lesson 26).
    pub visible: Vec<Vec<bool>>,
}

/// Parses rows of '1' and '0'.
fn parse_visible(rows: &[String]) -> Vec<Vec<bool>> {
    rows.iter().map(|r| r.chars().map(|c| c == '1').collect()).collect()
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
            visible: parse_visible(&z.visible),
        }
    }
}

impl ZoneView {
    /// Applies a delta. A delta for another room is ignored (it may arrive
    /// just before the new zone); returns whether it was applied.
    pub fn apply(&mut self, d: DeltaData) -> bool {
        if d.room != self.room {
            return false;
        }
        for e in d.added.into_iter().chain(d.moved) {
            self.entities.insert(e.id.clone(), e);
        }
        for id in &d.removed {
            self.entities.remove(id);
        }
        // Lesson 28: tiles seen for the first time, and the new field of view.
        for t in &d.tiles {
            let Some(code) = t.code.chars().next() else { continue };
            if let Some(cell) = self.rows.get_mut(t.y).and_then(|row| row.get_mut(t.x)) {
                *cell = code;
            }
            if !t.meaning.is_empty() {
                self.legend.insert(t.code.clone(), t.meaning.clone());
            }
        }
        if !d.visible.is_empty() {
            self.visible = parse_visible(&d.visible);
        }
        true
    }

    /// Where the player stands, if known.
    pub fn self_pos(&self) -> Option<(usize, usize)> {
        self.entities
            .get(&self.self_id)
            .map(|e| (e.x as usize, e.y as usize))
    }

    /// Whether the player sees (x, y) right now.
    pub fn is_visible(&self, x: usize, y: usize) -> bool {
        if self.visible.is_empty() {
            return true;
        }
        self.visible
            .get(y)
            .and_then(|row| row.get(x))
            .copied()
            .unwrap_or(false)
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
    if size <= view {
        return 0;
    }
    focus.saturating_sub(view / 2).min(size - view)
}
