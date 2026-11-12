# Lektion 40 – Kärnan

I lektion 41 ska din Rust-klient köras i webbläsaren. Det låter som en
omskrivning – men det mesta av klienten bryr sig inte om var den körs. Att
tolka en rad från servern, hålla zonen uppdaterad, välja tecken och färg för
en skog, ljuset från en fackla, en väg till en klickad ruta: inget av det
behöver en terminal, ett fönster, ett nätverk eller en klocka.

Så klienten delas i två. **Kärnan** – ett eget bibliotek, `mud-core` – får
allt som är likadant överallt. Terminalklienten, fönstret och (nästa lektion)
webbläsaren blir tunna skal runt den:

```
client/                      en Cargo-*workspace*
├── Cargo.toml               mud-client: terminalen och fönstret
├── src/                     app, scene, net, config, messages, gfx …
└── core/                    mud-core: protocol, zone, render, tileset,
    ├── Cargo.toml                     light, path – ingen I/O
    └── src/
```

## Mål

- Dela en crate i en **workspace** med ett eget bibliotek.
- Dra gränsen mellan ren logik och I/O – och håll den med ett test.
- Möt Rusts **orphan rule** (föräldralösa-regeln), och kom runt den.
- Skicka in tiden i stället för att läsa en klocka.

## Rust-koncept

### En workspace

```toml
[dependencies]
mud-core = { path = "core" }

[workspace]
members = ["core"]
```

Medlemmarna delar en `Cargo.lock` och en `target/`. `mud-client` beror på
`mud-core` som på vilken crate som helst, via sökvägen.

### Återexporter behåller de gamla sökvägarna

```rust
pub use mud_core::{protocol, render, tileset, zone};
```

i klientens `lib.rs` betyder att `crate::protocol::ServerMsg` och
`mud_client::zone::ZoneView` fortfarande fungerar – samma typer, som nu bor i kärnan.

### Orphan rule

Två saker slutar kompilera efter flytten:

```rust
impl fmt::Display for Status { ... }   // Display: std:s.  Status: kärnans.
impl From<Rgb> for Color { ... }       // From: std:s.  Rgb: kärnans.  Color: ratatuis.
```

Rust låter dig bara implementera ett trait för en typ om din crate äger
traitet *eller* typen. Annars skulle två crates kunna implementera samma sak
på två sätt, och ingen skulle veta vilket som gäller. Klienten äger ingetdera,
så de blir vanliga funktioner: `status_line(&Status)` och `color(Rgb)`.
(Kärnan *skulle* kunna implementera `Display` – men ordet "online" är
klientens, på kursens språk.)

### Tid som ett tal

Kärnan får inte anropa `Instant::now()` – i webbläsaren blir det panik. Ljus
som skimrar får tiden utifrån, som millisekunder:

```rust
pub fn zone_light(zone: &ZoneView, terrain: &str, x: usize, y: usize, t_ms: u64) -> Rgb
```

Terminalklienten skickar tiden sedan den startade; webbläsaren kommer att
skicka `requestAnimationFrame`s klocka; testerna skickar vad de vill.

## Kontraktet

- `client/core/` är ett paket `mud-core` (givet: `Cargo.toml` och `src/lib.rs`)
  som bara beror på `serde` och `serde_json`, och vars kod inte gör någon I/O
  och inte läser någon klocka (ingen `std::net`, `std::fs`, `std::thread`,
  `Instant::now`, …).
- Det har modulerna `protocol`, `zone`, `render`, `tileset` (flyttade som de
  är, utom `Display for Status`), `light` (den rena delen: `Light`, `TORCH`,
  `ambient`, `light_at`, `lights_for`, `shimmer`, `scale` – och nu `shade`,
  `FOG`, `FULL_LIGHT` och nya `zone_light`) och `path` (`Cell`, `walkable`,
  `STEPS`, `step_command`, `find_path` från `mouse.rs`).
- `zone_light(zone, terrain, x, y, t_ms)`: `FOG` om rutan inte syns, annars
  omgivningsljuset plus facklorna, och vatten skimrar med `t_ms`.
- Klienten beror på `mud-core`, återexporterar `protocol`, `render`, `tileset`
  och `zone`, och dess `light` och `mouse` återexporterar kärnans delar – de
  gamla sökvägarna fungerar fortfarande.
- `scene::status_line(&Status) -> String` (med `messages::ONLINE`) och
  `scene::cell_light` använder `zone_light`; `app::color(Rgb) -> Color`.

## Uppgift

1. Lägg till `mud-core = { path = "core" }` och `[workspace] members = ["core"]`
   i `client/Cargo.toml`.
2. Flytta `protocol.rs`, `zone.rs`, `render.rs` och `tileset.rs` till `core/src/`
   (`git mv` om du använder git), och ta bort `impl Display for Status`.
3. Dela `light.rs` (partiklarna och skakningen stannar) och `mouse.rs` (menyn
   och tooltipen stannar; vägarna flyttar till `core/src/path.rs`).
4. Skriv `zone_light`, och rätta det kompilatorn hittar: `crate::`-sökvägar i
   kärnan, `status_line`, `color`.

Kör `./mudschool run client` och `./mudschool run gfx`: allt ska vara som förut.

## Go ↔ Rust

| | Go | Rust |
|---|---|---|
| Flera paket, ett bygge | en modul med paket i mappar | en workspace med crates |
| Vem får lägga till metoder | bara paketet som definierar typen | craten som äger traitet eller typen (orphan rule) |
| Återexport | – (anroparna importerar det nya paketet) | `pub use` |

## Fundera på

- Vilka moduler skulle kunna flytta till kärnan, men gjorde det inte? Varför inte `messages.rs`?
- Orphan rule förbjuder `impl From<Rgb> for Color` i klienten. Var *skulle*
  den kunna bo?
- Varför är "ingen klocka" mer än en WebAssembly-detalj? (Tänk på testerna.)

## Utmaningar

- Flytta de delar av `App` som inte handlar om terminalen till kärnan: ett
  `GameState` som både terminalen och webbläsaren skulle kunna använda.
- `cargo build --target wasm32-unknown-unknown -p mud-core` – byggs den? Det
  borde den, redan nu.
