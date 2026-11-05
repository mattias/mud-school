# Lesson 19 – Combat in the client

The server fights round by round and sends a status after every round – with
your HP and, since lesson 18, `status.target`: who you are fighting and how
much health they have left. Now the client shows it: a **target panel** with a
health bar under the map, and a status bar that **flashes red** when you get
hit.

A flash is the first thing in the client that depends on *time*: red for 300
ms, then back to normal. Code that calls `Instant::now()` deep inside is hard
to test. The trick is the same as on the server: **inject the clock**.

```rust
app.on_server_at(msg, now);   // the tests pass any Instant they like
app.is_flashing(now)
draw_at(frame, &app, now)
// and for real use:
app.on_server(msg);           // = on_server_at(msg, Instant::now())
```

## Goals

- Add an optional nested struct to a protocol (`Option<Target>` with `#[serde(default)]`).
- Use more ratatui: nested layouts and the `Gauge` widget.
- Work with `Instant` and `Duration`, and inject time to make code testable.
- Change the event loop so that the screen updates even without events (`recv_timeout`).

## Rust concepts

### Time

```rust
use std::time::{Duration, Instant};
pub const FLASH: Duration = Duration::from_millis(300);

let until = now + FLASH;           // Instant + Duration = Instant
let flashing = now < until;
self.flash_until.is_some_and(|until| now < until)  // Option<Instant>
```

### Waking up without events

The main loop waited with `rx.recv()` until a key or a server message
arrived. A flash must also *end* without anything happening, so wait at most
50 ms:

```rust
match rx.recv_timeout(Duration::from_millis(50)) {
    Ok(ev) => { /* handle it */ }
    Err(RecvTimeoutError::Timeout) => continue,     // just redraw
    Err(RecvTimeoutError::Disconnected) => break,
}
```

## The contract

- `protocol.rs`: `pub struct Target { pub name: String, pub hp: i32, pub max_hp: i32 }`
  and `Status::target: Option<Target>` (`#[serde(default)]`).
- `app.rs`:
  - `pub const FLASH: Duration` (300 ms), `pub const FLASH_BG: Color` and
    `pub flash_until: Option<Instant>` in `App`.
  - `on_server_at(&mut self, msg, now)` – when a status has lower `hp` than
    the previous one, flash until `now + FLASH`. `on_server(msg)` calls it
    with `Instant::now()`.
  - `is_flashing(&self, now) -> bool`.
  - `pub fn draw_at(frame, app, now)`; `draw(frame, app)` calls it with
    `Instant::now()`. While flashing, the status bar's background is `FLASH_BG`.
  - With a target: a panel under the map with the target's name as the title
    and a `Gauge` labelled `hp/max_hp` (negative hp shows as 0).
- `main.rs`: `recv_timeout` as above.

## Task

Update `protocol.rs`, `app.rs` and `main.rs`. Then run the server from lesson
18 and fight the rat.

## Think about

- Why isn't `is_flashing` simply a `bool` field that a timer resets?
- What other things in a game client depend on time? (Animations, a blinking
  cursor, "x seconds ago" …) How would you test them?
- `Gauge::ratio` takes an `f64` from 0 to 1. What happens if the server sends
  `hp` > `max_hp`?

## Challenges

- Show the damage as a number that floats up and fades (store a list of
  effects with start times).
- Colour your own HP hearts yellow at half health and red below a quarter.
- Shake the map: offset it by one column for a few frames when you are hit.
