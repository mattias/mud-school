# Lektion 25 – Kameran

Sedan lektion 23 skickar servern zoner och deltor – och din klient har
artigt ignorerat dem (`#[serde(other)] Unknown`, lektion 11). Nu ritar den
dem: zonen fyller större delen av skärmen, loggen flyttar ner, och
piltangenterna går.

En zon är 31×13 rutor som standard men kan vara mycket större än terminalen.
Då visar klienten ett *fönster* av den – en **kamera** som följer spelaren och
aldrig visar det som ligger utanför zonen.

```
zon (40 bred)    ##########################################
kamera (10)                 [ ....@..... ]       fokus 20 -> börjar på 15
nära kanten                                 [ .......@.# ]  aldrig förbi slutet
```

## Mål

- Deserialisera nästlad JSON med listor av structs, och ett fält som heter
  `self` (ett nyckelord i Rust) med `#[serde(rename = "self")]`.
- Hålla tillstånd uppdaterat med stegvisa uppdateringar (deltor).
- Använda `BTreeMap` när ordningen spelar roll och `HashMap` när den inte gör det.
- Räkna säkert med tal utan tecken: `saturating_sub`, `min`.
- Lägga ut en mer komplex skärm med nästlade ratatui-layouter.

## Rust-koncept

### `self` som fältnamn

```rust
#[derive(Deserialize)]
pub struct ZoneData {
    // ...
    #[serde(rename = "self", default)]
    pub self_id: String,
}
```

### `From`

`impl From<ZoneData> for ZoneView` gör om det servern skickade till det
klienten sparar: raderna som `Vec<Vec<char>>` (så att `rows[y][x]` är ett
tecken, inte en byteposition i en UTF-8-sträng) och entiteterna i en map
efter ID.

### `saturating_sub`

`usize` kan inte vara negativt: `2 - 5` får panik i ett debug-bygge.
`2usize.saturating_sub(5)` är `0`. Det är precis vad en kamera vid vänsterkanten behöver:

```rust
focus.saturating_sub(view / 2).min(size - view)
```

### Match-villkor

Piltangenterna gör två saker nu – går när inmatningsraden är tom, bläddrar i
historiken medan du skriver:

```rust
KeyCode::Up | KeyCode::Down | KeyCode::Left | KeyCode::Right
    if self.input.is_empty() && self.history_pos.is_none() && !self.secret =>
{
    return walk_command(key.code).map(String::from);
}
KeyCode::Up if !self.secret => self.history_up(),
```

## Kontraktet

- `protocol.rs`: `Entity {id, kind, name, glyph, x, y}`, `ZoneData {room,
  width, height, rows, legend, entities, self_id}` och `DeltaData {room,
  added, moved, removed}` (listor som saknas = tomma); `ServerMsg::Zone { zone }`
  och `ServerMsg::Delta { delta }`.
- `zone.rs` (delvis given): `ZoneView` med `From<ZoneData>`, `apply(delta) ->
  bool` (en delta för ett annat rum ignoreras och ger `false`), `self_pos()`,
  `code_at`, `meaning_at`, `entity_at`; och `camera(focus, view, size) ->
  usize`: `focus` i mitten, begränsat till `0..=size-view`, och `0` när zonen får plats.
- `render.rs`: rutor för `floor`, `wall`, `grass`, `tree`, `rock`, `door`;
  `entity_cell(kind, glyph, is_self)`; `render_zone(view, x0, y0, w, h)`.
- `app.rs`: `App::zone: Option<ZoneView>`, sätts av zone-frames och
  uppdateras av deltor; med en zon är skärmen zonen överst och loggen under;
  med tom inmatningsrad returnerar piltangenterna `"north"`, `"south"`,
  `"west"`, `"east"` (och inget hamnar i loggen eller historiken); medan du
  skriver bläddrar Upp/Ner i historiken som förut.

## Uppgift

Protokollet först, sedan `zone.rs`, `render.rs` och appen. Kör servern från
lektion 24 och gå omkring med piltangenterna. Gör terminalen liten och titta på kameran.

## Fundera på

- Varför kan en delta komma för "fel" rum? (Tänk på ordningen på frames när
  du går ut genom en öppning.)
- `entities` är en `BTreeMap`, `legend` en `HashMap`. Varför skillnaden?
- `render_zone` frågar `entity_at` för varje ruta – den söker igenom alla
  entiteter. Hur många rutor och entiteter innan det spelar roll? Hur skulle du snabba upp det?
- Vad händer om servern skickar en delta innan någon zon alls?

## Utmaningar

- Diagonala steg: siffertangenterna 7, 9, 1, 3 (eller `y u b n` som i *NetHack*).
- En kant i en annan färg där zonen tar slut, när zonen är mindre än skärmen.
- Håll Skift + pil för att fortsätta gå tills något intressant händer ("springa").
