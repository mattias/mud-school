//! The client's state and how it is drawn.
//!
//! The architecture is "The Elm Architecture" in miniature:
//!   event (key / server message) -> update App -> draw App.
//! All logic lives in App and can be tested without a terminal or a network.

use ratatui::crossterm::event::{KeyCode, KeyEvent, KeyEventKind, KeyModifiers};
use ratatui::layout::{Constraint, Layout, Rect};
use ratatui::style::{Color, Modifier, Style};
use ratatui::text::{Line, Span};
use std::time::{Duration, Instant};

use ratatui::widgets::{Block, Borders, Gauge, Paragraph};
use ratatui::Frame;

use crate::messages;
use crate::protocol::{Kind, MapData, ServerMsg, Status, Target};
use crate::render::{render_map, render_zone, Cell, Rgb};
use crate::scene::{shade, FOG};
use crate::zone::{camera, ZoneView};

/// The maximum number of lines kept in the log.
pub const MAX_LOG: usize = 1000;

/// The maximum number of commands kept in the history.
pub const MAX_HISTORY: usize = 100;

/// How many lines PageUp/PageDown scroll.
pub const PAGE: usize = 10;

/// How long the status bar flashes when you lose health.
pub const FLASH: Duration = Duration::from_millis(300);

/// The status bar's background while it flashes.
pub const FLASH_BG: Color = Color::Rgb(150, 20, 20);

/// The status bar's normal background.
pub const STATUS_BG: Color = Color::Rgb(40, 40, 60);

impl From<Rgb> for Color {
    fn from(Rgb(r, g, b): Rgb) -> Color {
        Color::Rgb(r, g, b)
    }
}

/// One line in the log.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct LogLine {
    pub kind: Kind,
    pub text: String,
}

/// How a kind of text is drawn in the log.
pub fn style_for(kind: Kind) -> Style {
    match kind {
        Kind::Room => Style::default().fg(Color::Rgb(230, 230, 240)),
        Kind::Say => Style::default().fg(Color::Rgb(120, 200, 255)),
        Kind::Move => Style::default().fg(Color::Rgb(150, 150, 160)),
        Kind::Error => Style::default().fg(Color::Rgb(255, 110, 90)),
        Kind::Combat => Style::default()
            .fg(Color::Rgb(255, 190, 60))
            .add_modifier(Modifier::BOLD),
        Kind::Info | Kind::Other => Style::default().fg(Color::Rgb(200, 200, 200)),
    }
}

/// Whether a prompt from the server asks for something that should be hidden.
pub fn is_secret_prompt(prompt: &str) -> bool {
    matches!(prompt, "password" | "new_password")
}

/// What the input line shows: the text, or one * per character while secret.
pub fn shown_input(app: &App) -> String {
    if app.secret {
        "*".repeat(app.input.chars().count())
    } else {
        app.input.clone()
    }
}

/// The command an arrow key sends when the input line is empty.
pub fn walk_command(code: KeyCode) -> Option<&'static str> {
    match code {
        KeyCode::Up => Some("north"),
        KeyCode::Down => Some("south"),
        KeyCode::Left => Some("west"),
        KeyCode::Right => Some("east"),
        _ => None,
    }
}

/// Everything the client knows right now.
#[derive(Debug, Default)]
pub struct App {
    pub log: Vec<LogLine>,
    pub map: Option<MapData>,
    pub status: Option<Status>,
    pub input: String,
    pub quit: bool,
    /// Commands sent earlier, oldest first.
    pub history: Vec<String>,
    /// Where Up/Down is in the history (None = editing a new line).
    pub history_pos: Option<usize>,
    /// What was typed before Up was pressed, so Down can bring it back.
    pub draft: String,
    /// How many lines the log is scrolled up from the bottom (0 = newest).
    pub scroll: usize,
    /// Until when the status bar flashes (you just lost health).
    pub flash_until: Option<Instant>,
    /// True while the server waits for a password: the input is hidden and
    /// not saved anywhere.
    pub secret: bool,
    /// The zone the player is in (lesson 25).
    pub zone: Option<ZoneView>,
}

impl App {
    pub fn new() -> App {
        App::default()
    }

    /// Adds text to the log (one entry per line of the text).
    pub fn push_log(&mut self, kind: Kind, text: &str) {
        let before = self.log.len();
        self.log.extend(text.lines().map(|l| LogLine {
            kind,
            text: l.to_string(),
        }));
        // Keep the view still when the player has scrolled up.
        if self.scroll > 0 {
            self.scroll += self.log.len() - before;
        }
        if self.log.len() > MAX_LOG {
            let extra = self.log.len() - MAX_LOG;
            self.log.drain(..extra);
        }
        self.scroll = self.scroll.min(self.log.len());
    }

    /// Handles a message from the server.
    pub fn on_server(&mut self, msg: ServerMsg) {
        self.on_server_at(msg, Instant::now());
    }

    /// Like on_server, with the current time passed in (so tests control the clock).
    pub fn on_server_at(&mut self, msg: ServerMsg, now: Instant) {
        match msg {
            ServerMsg::Text { text, kind } => self.push_log(kind, &text),
            ServerMsg::Map { map } => self.map = Some(map),
            ServerMsg::Status { status } => {
                if let Some(old) = &self.status {
                    if status.hp < old.hp {
                        self.flash_until = Some(now + FLASH);
                    }
                }
                self.status = Some(status);
            }
            ServerMsg::Zone { zone } => self.zone = Some(ZoneView::from(zone)),
            ServerMsg::Delta { delta } => {
                if let Some(z) = &mut self.zone {
                    z.apply(delta);
                }
            }
            ServerMsg::Prompt { prompt } => {
                self.secret = is_secret_prompt(&prompt);
            }
            ServerMsg::Hello { .. } | ServerMsg::Unknown => {}
        }
    }

    /// Whether the status bar should flash at time now.
    pub fn is_flashing(&self, now: Instant) -> bool {
        self.flash_until.is_some_and(|until| now < until)
    }

    /// The server disconnected.
    pub fn on_disconnect(&mut self) {
        self.push_log(Kind::Error, messages::DISCONNECTED);
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
            // With an empty input line the arrow keys walk (lesson 25) ...
            KeyCode::Up | KeyCode::Down | KeyCode::Left | KeyCode::Right
                if self.input.is_empty() && self.history_pos.is_none() && !self.secret =>
            {
                return walk_command(key.code).map(String::from);
            }
            // ... while typing, Up/Down browse the history.
            KeyCode::Up if !self.secret => self.history_up(),
            KeyCode::Down if !self.secret => self.history_down(),
            KeyCode::PageUp => self.scroll = (self.scroll + PAGE).min(self.log.len()),
            KeyCode::PageDown => self.scroll = self.scroll.saturating_sub(PAGE),
            KeyCode::End => self.scroll = 0,
            KeyCode::Enter => {
                let line = std::mem::take(&mut self.input);
                self.scroll = 0;
                if self.secret {
                    // A password: not in the history, not in the log.
                    self.secret = false;
                    return Some(line);
                }
                if !line.is_empty() && self.history.last() != Some(&line) {
                    self.history.push(line.clone());
                    if self.history.len() > MAX_HISTORY {
                        self.history.remove(0);
                    }
                }
                self.history_pos = None;
                self.scroll = 0;
                self.push_log(Kind::Info, &format!("> {line}"));
                return Some(line);
            }
            _ => {}
        }
        None
    }

    fn history_up(&mut self) {
        let pos = match self.history_pos {
            None if self.history.is_empty() => return,
            None => {
                self.draft = std::mem::take(&mut self.input);
                self.history.len() - 1
            }
            Some(0) => 0,
            Some(p) => p - 1,
        };
        self.history_pos = Some(pos);
        self.input = self.history[pos].clone();
    }

    fn history_down(&mut self) {
        match self.history_pos {
            None => {}
            Some(p) if p + 1 < self.history.len() => {
                self.history_pos = Some(p + 1);
                self.input = self.history[p + 1].clone();
            }
            Some(_) => {
                self.history_pos = None;
                self.input = std::mem::take(&mut self.draft);
            }
        }
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
    draw_at(frame, app, Instant::now());
}

/// Like draw, with the current time passed in (so tests control the clock).
pub fn draw_at(frame: &mut Frame, app: &App, now: Instant) {
    let [status_area, main_area, input_area] = Layout::vertical([
        Constraint::Length(1),
        Constraint::Min(3),
        Constraint::Length(3),
    ])
    .areas(frame.area());

    let map_width = app.map.as_ref().map_or(0, |m| m.width as u16 + 2).max(15);
    let [left_area, side_area] =
        Layout::horizontal([Constraint::Min(10), Constraint::Length(map_width)]).areas(main_area);

    draw_status(frame, app, status_area, app.is_flashing(now));
    // With a zone: the zone on top, the log below. Without: only the log.
    if let Some(zone) = &app.zone {
        let log_height = (left_area.height / 3).max(4);
        let [zone_area, log_area] =
            Layout::vertical([Constraint::Min(3), Constraint::Length(log_height)]).areas(left_area);
        draw_zone(frame, zone, zone_area);
        draw_log(frame, app, log_area);
    } else {
        draw_log(frame, app, left_area);
    }
    let target = app.status.as_ref().and_then(|s| s.target.as_ref());
    if let Some(target) = target {
        let [map_area, target_area] =
            Layout::vertical([Constraint::Min(3), Constraint::Length(3)]).areas(side_area);
        draw_map(frame, app, map_area);
        draw_target(frame, target, target_area);
    } else {
        draw_map(frame, app, side_area);
    }

    let title = if app.secret {
        messages::TITLE_PASSWORD
    } else {
        messages::TITLE_INPUT
    };
    let input = Paragraph::new(format!("> {}", shown_input(app)))
        .block(Block::default().borders(Borders::ALL).title(title));
    frame.render_widget(input, input_area);
    frame.set_cursor_position((
        input_area.x + 3 + app.input.chars().count() as u16,
        input_area.y + 1,
    ));
}

fn draw_status(frame: &mut Frame, app: &App, area: Rect, flashing: bool) {
    let bg = if flashing { FLASH_BG } else { STATUS_BG };
    let bar = Style::default().bg(bg).fg(Color::White);
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
    let end = app.log.len().saturating_sub(app.scroll);
    let start = end.saturating_sub(visible);
    let lines: Vec<Line> = app.log[start..end]
        .iter()
        .map(|l| Line::styled(l.text.as_str(), style_for(l.kind)))
        .collect();
    let mut title = messages::TITLE_LOG.to_string();
    if app.scroll > 0 {
        title = format!("{title}↑{} ", app.scroll);
    }
    let log = Paragraph::new(lines).block(Block::default().borders(Borders::ALL).title(title));
    frame.render_widget(log, area);
}

fn draw_map(frame: &mut Frame, app: &App, area: Rect) {
    let lines: Vec<Line> = match &app.map {
        None => vec![Line::from(messages::NO_MAP)],
        Some(map) => render_map(map)
            .into_iter()
            .map(|row| row.into_iter().map(cell_span).collect::<Line>())
            .collect(),
    };
    let block = Block::default()
        .borders(Borders::ALL)
        .title(messages::TITLE_MAP);
    frame.render_widget(Paragraph::new(lines).block(block), area);
}

/// The zone, with the camera following the player.
fn draw_zone(frame: &mut Frame, zone: &ZoneView, area: Rect) {
    let block = Block::default().borders(Borders::ALL);
    let inner = block.inner(area);
    frame.render_widget(block, area);
    let (w, h) = (inner.width as usize, inner.height as usize);
    let (sx, sy) = zone.self_pos().unwrap_or((0, 0));
    let (x0, y0) = (camera(sx, w, zone.width), camera(sy, h, zone.height));
    let lines: Vec<Line> = render_zone(zone, x0, y0, w.min(zone.width), h.min(zone.height))
        .into_iter()
        .enumerate()
        .map(|(vy, row)| {
            row.into_iter()
                .enumerate()
                .map(|(vx, c)| {
                    // Explored but not visible right now: the fog of war (lesson 28).
                    if zone.is_visible(x0 + vx, y0 + vy) {
                        cell_span(c)
                    } else {
                        cell_span(Cell {
                            fg: shade(c.fg, FOG),
                            bg: c.bg.map(|bg| shade(bg, FOG)),
                            ..c
                        })
                    }
                })
                .collect::<Line>()
        })
        .collect();
    frame.render_widget(Paragraph::new(lines), inner);
}

/// A cell as a coloured span.
fn cell_span(c: Cell) -> Span<'static> {
    let mut style = Style::default().fg(c.fg.into());
    if let Some(bg) = c.bg {
        style = style.bg(bg.into());
    }
    Span::styled(c.glyph.to_string(), style)
}

/// The opponent's name and a health bar.
fn draw_target(frame: &mut Frame, target: &Target, area: Rect) {
    let ratio = if target.max_hp > 0 {
        (target.hp.max(0) as f64 / target.max_hp as f64).clamp(0.0, 1.0)
    } else {
        0.0
    };
    let gauge = Gauge::default()
        .block(
            Block::default()
                .borders(Borders::ALL)
                .title(format!(" {} ", target.name)),
        )
        .gauge_style(Style::default().fg(Color::Rgb(200, 40, 40)))
        .ratio(ratio)
        .label(format!("{}/{}", target.hp.max(0), target.max_hp));
    frame.render_widget(gauge, area);
}
