# Hints

## Moving the files

```
cd workspace/client
git mv src/protocol.rs src/zone.rs src/render.rs src/tileset.rs core/src/   # or plain mv
```

Then remove `pub mod protocol;` and the others from `src/lib.rs` and add
`pub use mud_core::{protocol, render, tileset, zone};`.

## The core's own paths

Inside the core, `crate::` means the core: `use crate::protocol::MapData;`
in `render.rs` still works. What breaks is what pointed at the client –
`crate::messages::ONLINE` in `protocol.rs` (that is the `Display` that goes).

## light.rs in two

The core gets everything above `Particle`, plus:

```rust
pub fn zone_light(zone: &ZoneView, terrain: &str, x: usize, y: usize, t_ms: u64) -> Rgb {
    if !zone.is_visible(x, y) {
        return FOG;
    }
    let light = light_at(x, y, &lights_for(zone), ambient(terrain));
    if zone.meaning_at(x, y) == Some("water") {
        return scale(light, shimmer(x, y, t_ms));
    }
    light
}
```

The client's `light.rs` starts with `pub use mud_core::light::*;` and keeps
`Particle`, `Particles` and `shake_offset`.

## scene.rs

`shade`, `FOG` and `FULL_LIGHT` move to the core; `pub use crate::light::{shade, FOG, FULL_LIGHT};`
keeps `crate::scene::shade` working for `app.rs` and `gfx.rs`.

## "only traits defined in the current crate can be implemented"

That is the orphan rule (error E0117). Replace the `impl` with a function and
the `.into()` calls with calls to it.
