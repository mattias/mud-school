# Lektion 19 – Strid i klienten

Servern slåss runda för runda och skickar en status efter varje runda – med
ditt HP och, sedan lektion 18, `status.target`: vem du slåss mot och hur mycket
hälsa den har kvar. Nu visar klienten det: en **målpanel** med en hälsomätare
under kartan, och en statusrad som **blinkar rött** när du blir träffad.

En blinkning är det första i klienten som beror på *tid*: rött i 300 ms, sedan
normalt igen. Kod som anropar `Instant::now()` djupt inne är svår att testa.
Knepet är samma som på servern: **injicera klockan**.

```rust
app.on_server_at(msg, now);   // testerna skickar vilken Instant de vill
app.is_flashing(now)
draw_at(frame, &app, now)
// och på riktigt:
app.on_server(msg);           // = on_server_at(msg, Instant::now())
```

## Mål

- Lägga till en valfri inbäddad struct i ett protokoll (`Option<Target>` med `#[serde(default)]`).
- Använda mer ratatui: nästlade layouter och widgeten `Gauge`.
- Arbeta med `Instant` och `Duration`, och injicera tid för att göra kod testbar.
- Ändra händelseloopen så att skärmen uppdateras även utan händelser (`recv_timeout`).

## Rust-koncept

### Tid

```rust
use std::time::{Duration, Instant};
pub const FLASH: Duration = Duration::from_millis(300);

let until = now + FLASH;           // Instant + Duration = Instant
let flashing = now < until;
self.flash_until.is_some_and(|until| now < until)  // Option<Instant>
```

### Vakna utan händelser

Huvudloopen väntade med `rx.recv()` tills en tangent eller ett servermeddelande
kom. En blinkning måste också *sluta* utan att något händer, så vänta högst 50 ms:

```rust
match rx.recv_timeout(Duration::from_millis(50)) {
    Ok(ev) => { /* hantera den */ }
    Err(RecvTimeoutError::Timeout) => continue,     // rita bara om
    Err(RecvTimeoutError::Disconnected) => break,
}
```

## Kontraktet

- `protocol.rs`: `pub struct Target { pub name: String, pub hp: i32, pub max_hp: i32 }`
  och `Status::target: Option<Target>` (`#[serde(default)]`).
- `app.rs`:
  - `pub const FLASH: Duration` (300 ms), `pub const FLASH_BG: Color` och
    `pub flash_until: Option<Instant>` i `App`.
  - `on_server_at(&mut self, msg, now)` – när en status har lägre `hp` än den
    förra, blinka till `now + FLASH`. `on_server(msg)` anropar den med
    `Instant::now()`.
  - `is_flashing(&self, now) -> bool`.
  - `pub fn draw_at(frame, app, now)`; `draw(frame, app)` anropar den med
    `Instant::now()`. Under blinkningen är statusradens bakgrund `FLASH_BG`.
  - Med ett mål: en panel under kartan med målets namn som rubrik och en
    `Gauge` med texten `hp/max_hp` (negativt hp visas som 0).
- `main.rs`: `recv_timeout` enligt ovan.

## Uppgift

Uppdatera `protocol.rs`, `app.rs` och `main.rs`. Kör sedan servern från
lektion 18 och slåss mot råttan.

## Fundera på

- Varför är `is_flashing` inte bara ett `bool`-fält som en timer nollställer?
- Vilka andra saker i en spelklient beror på tid? (Animationer, en blinkande
  markör, "för x sekunder sedan" …) Hur skulle du testa dem?
- `Gauge::ratio` tar ett `f64` från 0 till 1. Vad händer om servern skickar
  `hp` > `max_hp`?

## Utmaningar

- Visa skadan som en siffra som flyter uppåt och bleknar (spara en lista med
  effekter och deras starttider).
- Färga dina egna HP-hjärtan gula vid halv hälsa och röda under en fjärdedel.
- Skaka kartan: flytta den en kolumn i några bildrutor när du blir träffad.
