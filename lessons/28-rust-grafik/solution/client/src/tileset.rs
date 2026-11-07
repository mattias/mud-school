//! The tileset: one PNG with all the 16×16 pictures side by side (an
//! *atlas*). `assets/tiles.png` is the course's own (CC0); you can swap it for
//! any other with the same order – or change the order here.

/// The size of a tile in the PNG, in pixels.
pub const TILE_PX: f32 = 16.0;

/// How many tiles there are per row in the atlas.
pub const ATLAS_COLUMNS: u16 = 16;

/// The atlas index of the picture for a tile meaning.
pub fn tile_sprite(meaning: &str) -> Option<u16> {
    Some(match meaning {
        "floor" => 0,
        "wall" => 1,
        "grass" => 2,
        "tree" => 3,
        "water" => 4,
        "rock" => 5,
        "door" => 6,
        "stairs_up" => 7,
        "stairs_down" => 8,
        _ => return None,
    })
}

/// The atlas index of the picture for an entity.
pub fn entity_sprite(kind: &str, is_self: bool) -> u16 {
    match kind {
        _ if is_self => 9,
        "player" => 10,
        "creature" => 11,
        "item" => 12,
        _ => 13,
    }
}

/// Where tile `index` is in the atlas: (x, y, width, height) in pixels.
pub fn atlas_rect(index: u16) -> (f32, f32, f32, f32) {
    let col = (index % ATLAS_COLUMNS) as f32;
    let row = (index / ATLAS_COLUMNS) as f32;
    (col * TILE_PX, row * TILE_PX, TILE_PX, TILE_PX)
}
