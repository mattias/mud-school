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

// TODO(lektion 13): impl From<Rgb> for Color, så att `rgb.into()` ger en
// ratatui-färg (Color::Rgb(r, g, b)).

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
        // TODO(lektion 13): en rad per rad i texten, och högst MAX_LOG rader totalt.
        todo!()
    }

    /// Hanterar ett meddelande från servern.
    pub fn on_server(&mut self, msg: ServerMsg) {
        // TODO(lektion 13)
        todo!()
    }

    /// Servern kopplade ner.
    pub fn on_disconnect(&mut self) {
        self.push_log("*** Anslutningen till servern stängdes. Tryck Esc för att avsluta.");
    }

    /// Hanterar en tangenttryckning. Returnerar en rad att skicka till
    /// servern när spelaren trycker Enter.
    pub fn on_key(&mut self, key: KeyEvent) -> Option<String> {
        // TODO(lektion 13): Esc och Ctrl+C avslutar, tecken läggs till,
        // Backspace tar bort, Enter returnerar raden (och tömmer input).
        // Ignorera allt som inte är KeyEventKind::Press.
        todo!()
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
    // TODO(lektion 13): namn, rum, HP och antal online – gärna med en HP-mätare.
}

fn draw_log(frame: &mut Frame, app: &App, area: Rect) {
    // TODO(lektion 13): visa de SENASTE raderna som får plats, i ett Block med ram.
}

fn draw_map(frame: &mut Frame, app: &App, area: Rect) {
    // TODO(lektion 13): render_map(...) -> en Line per rad, en Span per cell,
    // med Style::default().fg(...) (och .bg(...) om cellen har bakgrund).
}
