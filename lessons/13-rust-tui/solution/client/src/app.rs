//! Klientens tillstånd och hur det ritas.
//!
//! Arkitekturen är "The Elm Architecture" i miniformat:
//!   händelse (tangent / servermeddelande) -> uppdatera App -> rita App.
//! All logik ligger i App och kan testas utan terminal och utan nätverk.

use ratatui::crossterm::event::{KeyCode, KeyEvent, KeyEventKind, KeyModifiers};
use ratatui::layout::{Constraint, Layout, Rect};
use ratatui::style::{Color, Modifier, Style};
use ratatui::text::{Line, Span};
use ratatui::widgets::{Block, Borders, Paragraph};
use ratatui::Frame;

use crate::protocol::{MapData, ServerMsg, Status};
use crate::render::{render_map, Rgb};

/// Max antal rader som sparas i loggen.
pub const MAX_LOG: usize = 1000;

impl From<Rgb> for Color {
    fn from(Rgb(r, g, b): Rgb) -> Color {
        Color::Rgb(r, g, b)
    }
}

/// Allt klienten vet just nu.
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

    /// Lägger till text i loggen (en rad per rad i texten).
    pub fn push_log(&mut self, text: &str) {
        self.log.extend(text.lines().map(str::to_string));
        if self.log.len() > MAX_LOG {
            let extra = self.log.len() - MAX_LOG;
            self.log.drain(..extra);
        }
    }

    /// Hanterar ett meddelande från servern.
    pub fn on_server(&mut self, msg: ServerMsg) {
        match msg {
            ServerMsg::Text { text } => self.push_log(&text),
            ServerMsg::Map { map } => self.map = Some(map),
            ServerMsg::Status { status } => self.status = Some(status),
            ServerMsg::Hello { .. } | ServerMsg::Unknown => {}
        }
    }

    /// Servern kopplade ner.
    pub fn on_disconnect(&mut self) {
        self.push_log("*** Anslutningen till servern stängdes. Tryck Esc för att avsluta.");
    }

    /// Hanterar en tangenttryckning. Returnerar en rad att skicka till
    /// servern när spelaren trycker Enter.
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

/// Ritar hela gränssnittet:
///
/// ```text
/// ┌ statusrad ───────────────────────────────┐
/// │ logg                       │   karta     │
/// │                            │             │
/// ├────────────────────────────┴─────────────┤
/// │ > inmatning                              │
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

    let input = Paragraph::new(format!("> {}", app.input))
        .block(Block::default().borders(Borders::ALL).title(" Kommando "));
    frame.render_widget(input, input_area);
    frame.set_cursor_position((
        input_area.x + 3 + app.input.chars().count() as u16,
        input_area.y + 1,
    ));
}

fn draw_status(frame: &mut Frame, app: &App, area: Rect) {
    let bar = Style::default().bg(Color::Rgb(40, 40, 60)).fg(Color::White);
    let line = match &app.status {
        None => Line::from(" Inte inloggad"),
        Some(s) => {
            let hearts = 10;
            let full = if s.max_hp > 0 {
                (s.hp.max(0) * hearts / s.max_hp) as usize
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
                Span::raw(format!(" HP {}/{} · {} online", s.hp, s.max_hp, s.online)),
            ])
        }
    };
    frame.render_widget(Paragraph::new(line).style(bar), area);
}

fn draw_log(frame: &mut Frame, app: &App, area: Rect) {
    let visible = area.height.saturating_sub(2) as usize; // minus ramen
    let start = app.log.len().saturating_sub(visible);
    let lines: Vec<Line> = app.log[start..]
        .iter()
        .map(|l| Line::from(l.as_str()))
        .collect();
    let log = Paragraph::new(lines).block(Block::default().borders(Borders::ALL).title(" Logg "));
    frame.render_widget(log, area);
}

fn draw_map(frame: &mut Frame, app: &App, area: Rect) {
    let lines: Vec<Line> = match &app.map {
        None => vec![Line::from("(ingen karta)")],
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
    let block = Block::default().borders(Borders::ALL).title(" Karta ");
    frame.render_widget(Paragraph::new(lines).block(block), area);
}
