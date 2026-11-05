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

impl From<Rgb> for Color {
    fn from(Rgb(r, g, b): Rgb) -> Color {
        Color::Rgb(r, g, b)
    }
}

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
        self.log.extend(text.lines().map(str::to_string));
        if self.log.len() > MAX_LOG {
            let extra = self.log.len() - MAX_LOG;
            self.log.drain(..extra);
        }
    }

    /// Handles a message from the server.
    pub fn on_server(&mut self, msg: ServerMsg) {
        match msg {
            ServerMsg::Text { text } => self.push_log(&text),
            ServerMsg::Map { map } => self.map = Some(map),
            ServerMsg::Status { status } => self.status = Some(status),
            ServerMsg::Hello { .. } | ServerMsg::Unknown => {}
        }
    }

    /// The server disconnected.
    pub fn on_disconnect(&mut self) {
        self.push_log(messages::DISCONNECTED);
    }

    /// Handles a key press. Returns a line to send to the server when the
    /// player presses Enter.
    pub fn on_key(&mut self, key: KeyEvent) -> Option<String> {
        if key.kind != KeyEventKind::Press {
            return None;
        }
        match key.code {
            KeyCode::Esc => self.quit = true,
            KeyCode::Char('c') if key.modifiers.contains(KeyModifiers::CONTROL) => self.quit = true,
            KeyCode::Char(c) => self.input.push(c),
            KeyCode::Backspace => {
                self.input.pop();
            }
            KeyCode::Enter => {
                let line = std::mem::take(&mut self.input);
                self.push_log(&format!("> {line}"));
                return Some(line);
            }
            _ => {}
        }
        None
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
    let bar = Style::default().bg(Color::Rgb(40, 40, 60)).fg(Color::White);
    let line = match &app.status {
        None => Line::from(messages::NOT_LOGGED_IN),
        Some(s) => {
            let hearts = 10;
            let full = if s.max_hp > 0 {
                (s.hp.max(0) * hearts / s.max_hp).min(hearts) as usize
            } else {
                0
            };
            Line::from(vec![
                Span::styled(format!(" {} ", s.name), bar.add_modifier(Modifier::BOLD)),
                Span::raw(format!("· {} · ", s.room)),
                Span::styled(
                    "♥".repeat(full),
                    Style::default().fg(Color::Rgb(230, 50, 60)),
                ),
                Span::styled(
                    "♥".repeat(hearts as usize - full),
                    Style::default().fg(Color::Rgb(80, 80, 90)),
                ),
                Span::raw(format!(
                    " HP {}/{} · {} {}",
                    s.hp,
                    s.max_hp,
                    s.online,
                    messages::ONLINE
                )),
            ])
        }
    };
    frame.render_widget(Paragraph::new(line).style(bar), area);
}

fn draw_log(frame: &mut Frame, app: &App, area: Rect) {
    let visible = area.height.saturating_sub(2) as usize; // minus the border
    let start = app.log.len().saturating_sub(visible);
    let lines: Vec<Line> = app.log[start..]
        .iter()
        .map(|l| Line::from(l.as_str()))
        .collect();
    let log = Paragraph::new(lines).block(
        Block::default()
            .borders(Borders::ALL)
            .title(messages::TITLE_LOG),
    );
    frame.render_widget(log, area);
}

fn draw_map(frame: &mut Frame, app: &App, area: Rect) {
    let lines: Vec<Line> = match &app.map {
        None => vec![Line::from(messages::NO_MAP)],
        Some(map) => render_map(map)
            .into_iter()
            .map(|row| {
                row.into_iter()
                    .map(|c| {
                        let mut style = Style::default().fg(c.fg.into());
                        if let Some(bg) = c.bg {
                            style = style.bg(bg.into());
                        }
                        Span::styled(c.glyph.to_string(), style)
                    })
                    .collect::<Line>()
            })
            .collect(),
    };
    let block = Block::default()
        .borders(Borders::ALL)
        .title(messages::TITLE_MAP);
    frame.render_widget(Paragraph::new(lines).block(block), area);
}
