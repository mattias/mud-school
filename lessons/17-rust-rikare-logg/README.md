# Lesson 17 – A richer log

Since lesson 15 the server says *what kind* of text it sends:
`{"type":"text","text":"Bertil says: hi","kind":"say"}`. Now the client puts
that to use: speech in blue, errors in red, the room description bright,
comings and goings in grey. On top of that, two things every MUD client has:
a **command history** (arrow up/down) and **scrolling** (PageUp/PageDown).

## Goals

- Extend an enum and a protocol in a backwards-compatible way (`#[serde(default)]`, `#[serde(other)]`).
- Use `#[derive(Default)]` with `#[default]` on an enum variant.
- Model small state machines (history position, scroll offset) and test them.
- Work with slices and indices safely: `saturating_sub`, `min`.

## Rust concepts

### An enum with a default

```rust
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum Kind {
    #[default]
    Info,          // Kind::default() == Kind::Info
    Room, Say, Move, Error, Combat,
    #[serde(other)]
    Other,         // "kind":"poetry" from a newer server
}
```

With `#[serde(default)] kind: Kind` on the text variant, a server from lesson
14 (which sends no `kind`) still works. That is the same compatibility rule
you followed on the server side.

### Safe index arithmetic

```rust
let end = app.log.len().saturating_sub(app.scroll);   // never below 0
let start = end.saturating_sub(visible);
let lines = &app.log[start..end];
```

`usize` cannot be negative: `3 - 5` panics in debug builds. `saturating_sub`
stops at 0 instead.

## The contract

- `protocol.rs`: `pub enum Kind { Info, Room, Say, Move, Error, Combat, Other }`
  (default `Info`, lowercase in JSON, unknown → `Other`), and
  `ServerMsg::Text { text: String, kind: Kind }` (`kind` defaults to `Info`).
  `parse_line` of plain text gives `Kind::Info`.
- `app.rs`:
  - `pub struct LogLine { pub kind: Kind, pub text: String }`, and
    `App::log: Vec<LogLine>`; `push_log(&mut self, kind: Kind, text: &str)`.
  - `pub fn style_for(kind: Kind) -> Style` – a different style per kind
    (`Room`, `Say`, `Move`, `Error`, `Combat`); `Other` looks like `Info`.
    The log draws every line with its style.
  - History: `pub history: Vec<String>` (oldest first, at most
    `MAX_HISTORY`, the same command twice in a row is stored once). Up shows
    older commands, Down newer ones, and after the newest you get back what
    you were typing.
  - Scrolling: `pub scroll: usize` (lines up from the bottom), `pub const PAGE: usize = 10`.
    PageUp/PageDown move by `PAGE`, End and Enter go back to the bottom.
    While scrolled up, new messages don't move the view.

## Task

1. Update `protocol.rs` and `app.rs` as above. (`main.rs` needs a tiny change
   too: `push_log` now takes a kind.)
2. Show in the log's title that you are scrolled up, e.g. `Log ↑10`.

## Try it yourself

Run the server from lesson 16 and your client. Walk around, talk, type
nonsense – the colours should tell the story. Press Up a few times.

## Think about

- Why is `Kind` `Copy` but `LogLine` isn't?
- What happens with `Kind::Other` if a later server version adds `"kind":"whisper"`?
- History and scroll are tested without a terminal. What else in the client
  could be tested that way?

## Challenges

- Highlight your own name in the log.
- Tab completion of commands (`lo<Tab>` → `look`).
- Save the history to a file when the client quits and load it at start.
