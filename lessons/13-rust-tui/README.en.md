# Lesson 13 – A real user interface

Time to leave the "print lines and hope" stage. Now we build a client that
looks like a game: a status bar at the top, the log on the left, the map on
the right and an input line at the bottom – all redrawn many times a second.

```
 Anna · Mudby Square · ♥♥♥♥♥♥♥♥♥♥ HP 10/10 · 2 online
┌ Log ──────────────────────────────────┐┌ Map ────────┐
│Mudby Square                           ││      ■      │
│A cobbled square with an old well in   ││      │      │
│Exits: north, east, south, west        ││    ■─@─■    │
│Bertil says: hi!                       ││      │      │
│> say hi yourself                      ││      #      │
└───────────────────────────────────────┘└─────────────┘
┌ Command ─────────────────────────────────────────────┐
│> n_                                                  │
└──────────────────────────────────────────────────────┘
```

## Goals

- Use **ratatui**: layouts, widgets and immediate-mode rendering.
- Implement `From` to convert between types (`rgb.into()`).
- Combine keyboard and server into **one** event loop with an `enum Event`.
- Test a user interface without a terminal using `TestBackend`.

## Architecture

```
 keyboard thread ──Event::Key──────┐
                                   ├──► channel ──► main loop: update App ──► draw App
 server thread ────Event::Server───┘
```

This is sometimes called *The Elm Architecture*: all logic lives in `App`,
which is changed by events and drawn by a pure function `draw(frame, &app)`.
That means almost the whole client can be tested without a network or a
terminal – look in `mudschool_13.rs`.

**Go ↔ Rust:** the channel with several senders (`tx.clone()`) is exactly like
several goroutines sending to the same `chan` in Go.

## Rust concepts

### Immediate-mode rendering

ratatui draws **the whole screen every time** from your state. There are no
widgets that "live" between frames – you build them again in `draw`. ratatui
then compares with the previous frame and only sends the changes to the terminal.

```rust
let [top, middle, bottom] = Layout::vertical([
    Constraint::Length(1),   // exactly one row
    Constraint::Min(3),      // the rest
    Constraint::Length(3),
]).areas(frame.area());

frame.render_widget(Paragraph::new("hi").block(Block::bordered().title(" Log ")), middle);
```

Text is built from `Line`s made of `Span`s, and every `Span` has a `Style`:

```rust
Line::from(vec![
    Span::styled("♣", Style::default().fg(Color::Rgb(40, 170, 70))),
    Span::raw(" forest"),
])
```

### From and Into

```rust
impl From<Rgb> for Color {
    fn from(Rgb(r, g, b): Rgb) -> Color { Color::Rgb(r, g, b) }
}
let c: Color = rgb.into();   // you get Into for free!
```

### std::mem::take

`std::mem::take(&mut self.input)` takes the value out and leaves an empty
`String` behind – a common trick when you want to *move out* of a `&mut`.

## Task

1. `cargo add ratatui` (it brings `crossterm` along, which you reach as
   `ratatui::crossterm`).
2. **`src/app.rs`** (new file): fill in `From<Rgb> for Color`, `push_log`,
   `on_server`, `on_key`, `draw_status`, `draw_log` and `draw_map`. Titles and
   other texts are in `messages.rs` (`messages::TITLE_LOG`, ...). Add
   `pub mod app;` to `lib.rs`.
3. **`src/main.rs`**: rewrite the client:
   - an `enum Event { Key(KeyEvent), Server(ServerMsg), Disconnected }`,
   - a thread that reads the server and sends `Event::Server` (and finally
     `Event::Disconnected`), and a thread that reads the keyboard with
     `ratatui::crossterm::event::read()` and sends `Event::Key`,
   - `let mut terminal = ratatui::init();`, then a loop: draw, wait for an
     event, update `App`, send the line to the server if `on_key` returned one,
   - finish with `ratatui::restore();` – **even** if something went wrong,
     otherwise the terminal is left in "raw mode" (type `reset` if that happens).

## Try it yourself

```
./mudschool run server
./mudschool run client
```

Open another client (or telnet) and talk to yourself. Walk around and watch
the map and the status bar update.

## Think about

- Why do we read the keyboard in a thread of its own instead of in the main loop?
- `draw` takes `&App`, not `&mut App`. Why is that a good idea?
- What happens if the server sends 1000 lines at once? (Look at
  `rx.try_iter()` in the solution.)

## Challenges

- Colour the log: `Anna says:` in one colour, `You can't go that way.` in
  another. Maybe the server should send *what kind* of text it is? (That is
  exactly the kind of protocol extension we make in phase 3.)
- History: arrow up/down browse earlier commands.
- Scroll the log with PageUp/PageDown.
- Pop up a help window (`Clear` + `Paragraph` on top of everything else) with F1.
