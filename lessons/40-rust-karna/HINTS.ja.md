# ヒント

## ファイルを移す

```
cd workspace/client
git mv src/protocol.rs src/zone.rs src/render.rs src/tileset.rs core/src/   # または普通の mv
```

それから `src/lib.rs` から `pub mod protocol;` などを消し、
`pub use mud_core::{protocol, render, tileset, zone};` を足しましょう。

## コア自身のパス

コアの中では `crate::` はコアを指します。`render.rs` の `use crate::protocol::MapData;` は
そのまま動きます。壊れるのはクライアントを指していたもの、`protocol.rs` の
`crate::messages::ONLINE` です（なくなる `Display` の中）。

## light.rs を 2 つに

コアは `Particle` より上の全部と、次のものを持ちます:

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

クライアントの `light.rs` は `pub use mud_core::light::*;` で始め、`Particle`、
`Particles`、`shake_offset` を残します。

## scene.rs

`shade`、`FOG`、`FULL_LIGHT` はコアへ。`pub use crate::light::{shade, FOG, FULL_LIGHT};` で、
`app.rs` と `gfx.rs` の `crate::scene::shade` はそのまま使えます。

## "only traits defined in the current crate can be implemented"

それが孤児ルール（エラー E0117）です。`impl` を関数に、`.into()` をその関数の呼び出しに
置き換えましょう。
