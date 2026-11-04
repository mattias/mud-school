// Den här filen hör till kursverktyget och skrivs över av `mudschool check`.

use std::collections::HashMap;

use ratatui::backend::TestBackend;
use ratatui::buffer::Buffer;
use ratatui::crossterm::event::{KeyCode, KeyEvent, KeyModifiers};
use ratatui::style::Color;
use ratatui::Terminal;

use mud_client::app::{draw, App, MAX_LOG};
use mud_client::protocol::{MapData, ServerMsg, Status};
use mud_client::render::tile;

fn key(code: KeyCode) -> KeyEvent {
    KeyEvent::new(code, KeyModifiers::NONE)
}

fn type_text(app: &mut App, text: &str) {
    for c in text.chars() {
        assert_eq!(app.on_key(key(KeyCode::Char(c))), None);
    }
}

fn text(s: &str) -> ServerMsg {
    ServerMsg::Text {
        text: s.to_string(),
    }
}

fn sample_map() -> MapData {
    MapData {
        width: 3,
        height: 1,
        rows: vec!["T-@".into()],
        legend: HashMap::from([
            ("T".to_string(), "forest".to_string()),
            ("-".to_string(), "path_ew".to_string()),
            ("@".to_string(), "self".to_string()),
        ]),
    }
}

fn sample_status() -> Status {
    Status {
        name: "Anna".into(),
        room: "Torget".into(),
        terrain: "town".into(),
        hp: 7,
        max_hp: 10,
        online: 2,
    }
}

fn render(app: &App, w: u16, h: u16) -> Buffer {
    let mut terminal = Terminal::new(TestBackend::new(w, h)).unwrap();
    terminal.draw(|f| draw(f, app)).unwrap();
    terminal.backend().buffer().clone()
}

fn screen(buf: &Buffer) -> String {
    let area = buf.area;
    let mut s = String::new();
    for y in 0..area.height {
        for x in 0..area.width {
            s.push_str(buf[(x, y)].symbol());
        }
        s.push('\n');
    }
    s
}

#[test]
fn typing_and_enter() {
    let mut app = App::new();
    type_text(&mut app, "sägg");
    app.on_key(key(KeyCode::Backspace));
    assert_eq!(app.input, "säg");
    type_text(&mut app, " hej");
    assert_eq!(app.on_key(key(KeyCode::Enter)), Some("säg hej".to_string()));
    assert_eq!(app.input, "", "inmatningen ska tömmas efter Enter");
    assert!(!app.quit);
}

#[test]
fn escape_and_ctrl_c_quit() {
    let mut app = App::new();
    app.on_key(key(KeyCode::Esc));
    assert!(app.quit, "Esc ska avsluta");

    let mut app = App::new();
    app.on_key(KeyEvent::new(KeyCode::Char('c'), KeyModifiers::CONTROL));
    assert!(app.quit, "Ctrl+C ska avsluta");
    assert_eq!(app.input, "", "Ctrl+C ska inte skriva ett 'c'");
}

#[test]
fn server_messages_update_state() {
    let mut app = App::new();
    app.on_server(text("Torget\nEtt torg."));
    app.on_server(ServerMsg::Map { map: sample_map() });
    app.on_server(ServerMsg::Status {
        status: sample_status(),
    });
    app.on_server(ServerMsg::Unknown);
    assert!(
        app.log
            .ends_with(&["Torget".to_string(), "Ett torg.".to_string()]),
        "text med \\n ska bli flera rader i loggen, loggen är {:?}",
        app.log
    );
    assert_eq!(app.map, Some(sample_map()));
    assert_eq!(app.status, Some(sample_status()));
}

#[test]
fn log_is_capped() {
    let mut app = App::new();
    for i in 0..MAX_LOG + 50 {
        app.on_server(text(&format!("rad {i}")));
    }
    assert_eq!(app.log.len(), MAX_LOG);
    assert_eq!(app.log.last().unwrap(), &format!("rad {}", MAX_LOG + 49));
}

#[test]
fn draws_status_log_map_and_input() {
    let mut app = App::new();
    app.on_server(text("Välkommen, Anna!"));
    app.on_server(ServerMsg::Map { map: sample_map() });
    app.on_server(ServerMsg::Status {
        status: sample_status(),
    });
    type_text(&mut app, "look");

    let buf = render(&app, 80, 20);
    let s = screen(&buf);
    for want in ["Anna", "Torget", "7/10", "Välkommen, Anna!", "look"] {
        assert!(s.contains(want), "skärmen ska visa {want:?}:\n{s}");
    }
    let forest = tile("forest").glyph.to_string();
    assert!(
        s.contains(&forest),
        "kartan ska ritas med tile-tecken ({forest}):\n{s}"
    );

    // '@' ska ha samma färg som tile("self").
    let at = buf
        .content()
        .iter()
        .find(|c| c.symbol() == "@")
        .expect("spelaren '@' ska synas på kartan");
    let want = tile("self").fg;
    assert_eq!(
        at.fg,
        Color::Rgb(want.0, want.1, want.2),
        "'@' ska ritas i rätt färg"
    );
}

#[test]
fn log_scrolls_to_the_newest_lines() {
    let mut app = App::new();
    for i in 0..100 {
        app.on_server(text(&format!("meddelande nummer {i}")));
    }
    let s = screen(&render(&app, 80, 20));
    assert!(
        s.contains("meddelande nummer 99"),
        "senaste raden ska synas:\n{s}"
    );
    assert!(
        !s.contains("meddelande nummer 0 "),
        "äldsta raden ska ha scrollat bort:\n{s}"
    );
}

#[test]
fn rgb_converts_to_ratatui_color() {
    let c: Color = mud_client::render::Rgb(1, 2, 3).into();
    assert_eq!(c, Color::Rgb(1, 2, 3));
}
