# Tips

## Att flytta filerna

```
cd workspace/client
git mv src/protocol.rs src/zone.rs src/render.rs src/tileset.rs core/src/   # eller vanlig mv
```

Ta sedan bort `pub mod protocol;` och de andra ur `src/lib.rs` och lägg till
`pub use mud_core::{protocol, render, tileset, zone};`.

## Kärnans egna sökvägar

I kärnan betyder `crate::` kärnan: `use crate::protocol::MapData;` i
`render.rs` fungerar fortfarande. Det som går sönder är det som pekade på
klienten – `crate::messages::ONLINE` i `protocol.rs` (det är `Display` som försvinner).

## light.rs i två delar

Kärnan får allt ovanför `Particle`, och dessutom:

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

Klientens `light.rs` börjar med `pub use mud_core::light::*;` och behåller
`Particle`, `Particles` och `shake_offset`.

## scene.rs

`shade`, `FOG` och `FULL_LIGHT` flyttar till kärnan; `pub use crate::light::{shade, FOG, FULL_LIGHT};`
gör att `crate::scene::shade` fortfarande fungerar för `app.rs` och `gfx.rs`.

## "only traits defined in the current crate can be implemented"

Det är orphan rule (fel E0117). Byt ut `impl`-blocket mot en funktion och
`.into()`-anropen mot anrop till den.
