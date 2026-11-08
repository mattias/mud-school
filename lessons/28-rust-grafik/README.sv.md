# Lektion 28 – En grafisk klient

I lektion 10 lärde sig servern att skicka *betydelse* – `"forest"`,
`"wall"`, `"creature"` – i stället för utseende. Terminalklienten gjorde om
betydelser till färgade tecken. Nu lägger du till en andra klient: ett eget
**fönster**, med bilder – och resten av klienten (protokollet, nätverket,
App med sina tangenter och sin historik) förblir **exakt** densamma.

```
servern  ──JSON──>  protocol.rs  ──>  App  ──>  build_scene  ──>  Scene
                                                                   │
                                 ┌─────────────── Renderer ────────┴────────┐
                                 │                    │                     │
                           TextRenderer          terminalen            GfxRenderer
                           (tester)              (ratatui)             (ett fönster, tiles)
```

Fönstret använder **macroquad**, ett litet spelbibliotek, och ett
**tileset**: en PNG med alla 16×16-bilder bredvid varandra (en *atlas*).
Kursen ritar ett eget (`client/assets/tiles.png`, CC0, gjort av
`tools/tileset/main.go`) – samma idé som tileseten till *Dwarf Fortress*.
Byt gärna till ett annat.

## Mål

- Designa en **trait** som döljer *hur* något ritas – och implementera den mer än en gång.
- Skilja på *vad* som är på skärmen (en `Scene`, ren data) och *hur* det
  ritas. Ren data är lätt att testa.
- Använda **Cargo-features** och valfria beroenden, och ett andra program.
- Rita sprites från en atlas: index → rektangel i PNG:en.

## Rust-koncept

### En trait med flera implementationer

```rust
pub trait Renderer {
    fn cell(&mut self, x: u16, y: u16, cell: &SceneCell);
    fn text(&mut self, x: u16, y: u16, text: &str, fg: Rgb);
}

pub fn draw_scene(scene: &Scene, r: &mut impl Renderer) { ... }
```

`draw_scene` vet inte om den ritar tecken i ett rutnät eller texturer i ett
fönster. `impl Renderer` i argumentet är *statisk dispatch*: kompilatorn gör
en kopia av `draw_scene` per renderer-typ.

### Cargo-features

Grafikbiblioteket är stort, och terminalklienten behöver det inte:

```toml
macroquad = { version = "0.4", optional = true }

[features]
gfx = ["dep:macroquad"]

[[bin]]
name = "mud-gfx"
path = "gfx/main.rs"
required-features = ["gfx"]
```

Programmet ligger i `gfx/main.rs`, inte i `src/bin/`: Cargo hittar program i
`src/bin/` på egen hand – och skulle försöka bygga det här innan du har lagt
till macroquad. Posten `[[bin]]` talar om för Cargo var det finns.

`cargo test` och `cargo run` bygger inte macroquad alls. `cargo run
--release --features gfx --bin mud-gfx` gör det – och i koden finns
`#[cfg(feature = "gfx")] pub mod gfx;` bara då. Lägg till
`default-run = "mud-client"` under `[package]` så att `cargo run` fortfarande
vet vilket program du menar.

### Krigets dimma, äntligen

Sedan lektion 26 har deltor `tiles` (nyupptäckta rutor) och `visible`
(synfältet) – och din klient har ignorerat dem. Nu lägger `ZoneView::apply`
in rutorna i zonen och sparar `visible`; rutor du har sett men inte ser just
nu ritas mörkare (`shade(färg, FOG)`), i fönstret *och* i terminalen.

## Kontraktet

- `protocol.rs`: `ZoneData::visible: Vec<String>`; `TileChange {x, y, code,
  meaning}`; `DeltaData::tiles` och `DeltaData::visible` (saknas = tomma).
- `zone.rs`: `ZoneView::visible: Vec<Vec<bool>>` (från zone-framen),
  `is_visible(x, y)` (sant överallt när servern inte skickade `visible`);
  `apply` lägger in `tiles` i `rows` och deras betydelser i `legend`, och
  ersätter `visible` när deltan har ett.
- `tileset.rs`: `TILE_PX = 16.0`, `ATLAS_COLUMNS = 16`; `tile_sprite(meaning)`:
  floor 0, wall 1, grass 2, tree 3, water 4, rock 5, door 6, stairs_up 7,
  stairs_down 8, annars `None`; `entity_sprite(kind, is_self)`: du själv 9,
  spelare 10, varelse 11, sak 12, annars 13; `atlas_rect(index)` → `(x, y,
  w, h)` i pixlar (16 rutor per rad).
- `scene.rs`: `SceneCell {x, y, glyph, fg, bg, tile, entity, light}`, `Scene
  {width, height, cells, status, log, input}`, `LOG_LINES = 5`, `FULL_LIGHT`,
  `FOG`, `shade(färg, ljus)` (varje komponent × ljus / 255),
  `zone_rows(height)` = `height - 2 - LOG_LINES`, och:
  - `build_scene(app, width, height)` – `status` (statusradens text, eller
    `NOT_LOGGED_IN` trimmad), de sista `LOG_LINES` loggraderna, `input` =
    `"> "` + inmatningen som den visas; med en zon: en cell per **utforskad**
    ruta i kamerans vy (kameran från lektion 25, `width × zone_rows(height)`),
    på vykoordinater, med tecken och färger från `render_zone`, `tile` och
    `entity` från tilesetet, och `light` = `FULL_LIGHT` om den syns, annars `FOG`;
  - `draw_scene(scene, renderer)` – statusen på rad 0, cellerna en rad ner
    (`y + 1`), loggen under zonen, inmatningen på sista raden;
  - `TextRenderer` – ritar i ett rutnät av tecken; `lines()`.
- `gfx.rs` och `gfx/main.rs` (givna): fönster-renderern och programmet.
- `Cargo.toml`: featuren, det valfria beroendet, det andra programmet,
  `default-run`. `assets/tiles.png` är given.
- Terminalklienten gör också utforskade men inte synliga rutor mörkare.

## Uppgift

1. Protokollet och `zone.rs` (tiles, visible).
2. `tileset.rs` (`atlas_rect`) och `scene.rs`.
3. `Cargo.toml`. Kör sedan: `./mudschool run server` och
   `./mudschool run gfx --name Anna`. Första bygget tar en stund.

På Linux kan du behöva utvecklingspaketen för X11 och OpenGL (Ubuntu:
`sudo apt install libx11-dev libxi-dev libgl1-mesa-dev`); Windows och macOS
behöver inget extra.

Fönstret använder macroquads inbyggda typsnitt, som bara har latinska
tecken: japanska (och å, ä, ö) i loggen blir rutor. Se utmaningarna.

## Fundera på

- Varför är en `Scene` (ren data) lättare att testa än kod som ritar direkt?
- `&mut impl Renderer` mot `&mut dyn Renderer`: vad är skillnaden? När
  skulle du behöva `dyn`?
- Fönstret ritar om allt 60 gånger i sekunden; terminalen bara när något
  händer. Varför är det okej för ett spel?
- Vad skulle krävas för att visa *både* terminalen och fönstret samtidigt?

## Utmaningar

- Ett typsnitt med japanska och svenska tecken: `load_ttf_font` och en
  inställning `font = "..."` i `mudclient.toml` (t.ex. *Noto Sans JP* eller *UDEV Gothic*).
- Zooma med mushjulet (ändra `scale`).
- Ett annat tileset: Kenneys *1-Bit Pack* (CC0) – skriv en ny `tile_sprite`
  för det, eller låt `mudclient.toml` välja atlasfilen.
- Organisera klienten som en **Cargo workspace**: `mud-core` (protokoll,
  App, scen), `mud-tui` och `mud-gfx`.
