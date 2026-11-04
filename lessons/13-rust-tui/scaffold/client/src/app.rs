//! The client's state and how it is drawn.
//!
//! The architecture is "The Elm Architecture" in miniature:
//!   event (key / server message) -> update App -> draw App.
//! All logic lives in App and can be tested without a terminal or a network.

use ratatui::crossterm::event::{KeyCode, KeyEvent, KeyEventKind, KeyModifiers};
use ratatui::layout::{Constraint, Layout, Rect};
use ratatui::style::{Color, Modifier, Style};
use ratatui::text::{Line, Span};
use ratatui::widgets::{Block, Borders, Paragraph};
use ratatui::Frame;

use crate::messages;
use crate::protocol::{MapData, ServerMsg, Status};
use crate::render::{render_map, Rgb};

/// The maximum number of lines kept in the log.
pub const MAX_LOG: usize = 1000;

// TODO(lesson 13): impl From<Rgb> for Color, so that `rgb.into()` gives a
// ratatui colour (Color::Rgb(r, g, b)).

/// Everything the client knows right now.
#[derive(Debug, Default)]
pub struct App {
    pub log: Vec<String>,
    pub map: Option<MapData>,
    pub status: Option<Status>,
    pub input: String,
    pub quit: bool,
}

impl App {
    pub fn new() -> App {
        App::default()
    }

    /// Adds text to the log (one entry per line of the text).
    pub fn push_log(&mut self, text: &str) {
        // TODO(lesson 13): one entry per line of the text, and at most MAX_LOG entries in total.
        todo!()
    }

    /// Handles a message from the server.
    pub fn on_server(&mut self, msg: ServerMsg) {
        // TODO(lesson 13)
        todo!()
    }

    /// The server disconnected.
    pub fn on_disconnect(&mut self) {
        self.push_log(messages::DISCONNECTED);
    }

    /// Handles a key press. Returns a line to send to the server when the
    /// player presses Enter.
    pub fn on_key(&mut self, key: KeyEvent) -> Option<String> {
        // TODO(lesson 13): Esc and Ctrl+C quit, characters are appended,
        // Backspace deletes, Enter returns the line (and empties input).
        // Ignore everything that is not KeyEventKind::Press.
        todo!()
    }
}

/// Draws the whole user interface:
///
/// ```text
/// ┌ status bar ──────────────────────────────┐
/// │ log                        │   map       │
/// │                            │             │
/// ├────────────────────────────┴─────────────┤
/// │ > input                                  │
/// └──────────────────────────────────────────┘
/// ```
pub fn draw(frame: &mut Frame, app: &App) {
    let [status_area, main_area, input_area] = Layout::vertical([
        Constraint::Length(1),
        Constraint::Min(3),
        Constraint::Length(3),
    ])
    .areas(frame.area());

    let map_width = app.map.as_ref().map_or(0, |m| m.width as u16 + 2).max(15);
    let [log_area, map_area] =
        Layout::horizontal([Constraint::Min(10), Constraint::Length(map_width)]).areas(main_area);

    draw_status(frame, app, status_area);
    draw_log(frame, app, log_area);
    draw_map(frame, app, map_area);

    let input = Paragraph::new(format!("> {}", app.input)).block(
        Block::default()
            .borders(Borders::ALL)
            .title(messages::TITLE_INPUT),
    );
    frame.render_widget(input, input_area);
    frame.set_cursor_position((
        input_area.x + 3 + app.input.chars().count() as u16,
        input_area.y + 1,
    ));
}

fn draw_status(frame: &mut Frame, app: &App, area: Rect) {
    // TODO(lesson 13): name, room, HP and number online – preferably with an HP meter.
}

fn draw_log(frame: &mut Frame, app: &App, area: Rect) {
    // TODO(lesson 13): show the NEWEST lines that fit, in a Block with a border.
}

fn draw_map(frame: &mut Frame, app: &App, area: Rect) {
    // TODO(lesson 13): render_map(...) -> one Line per row, one Span per cell,
    // with Style::default().fg(...) (and .bg(...) if the cell has a background).
}
