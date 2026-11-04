// This file belongs to the course tool and is overwritten by `mudschool check`.

mod mudschool_common;

use std::collections::HashMap;

use ratatui::backend::TestBackend;
use ratatui::buffer::Buffer;
use ratatui::crossterm::event::{KeyCode, KeyEvent, KeyModifiers};
use ratatui::style::Color;
use ratatui::Terminal;

use mud_client::app::{draw, App, MAX_LOG};
use mud_client::protocol::{MapData, ServerMsg, Status};
use mud_client::render::tile;
use mudschool_common::l;

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
        room: "Square".into(),
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
    type_text(&mut app, "sayy");
    app.on_key(key(KeyCode::Backspace));
    assert_eq!(app.input, "say");
    type_text(&mut app, " hi");
    assert_eq!(app.on_key(key(KeyCode::Enter)), Some("say hi".to_string()));
    assert_eq!(
        app.input,
        "",
        "{}",
        l(
            "inmatningen ska tömmas efter Enter",
            "the input should be emptied after Enter",
            "Enter の後は入力を空にするべきです"
        )
    );
    assert!(!app.quit);
}

#[test]
fn escape_and_ctrl_c_quit() {
    let mut app = App::new();
    app.on_key(key(KeyCode::Esc));
    assert!(
        app.quit,
        "{}",
        l(
            "Esc ska avsluta",
            "Esc should quit",
            "Esc で終了するべきです"
        )
    );

    let mut app = App::new();
    app.on_key(KeyEvent::new(KeyCode::Char('c'), KeyModifiers::CONTROL));
    assert!(
        app.quit,
        "{}",
        l(
            "Ctrl+C ska avsluta",
            "Ctrl+C should quit",
            "Ctrl+C で終了するべきです"
        )
    );
    assert_eq!(
        app.input,
        "",
        "{}",
        l(
            "Ctrl+C ska inte skriva ett 'c'",
            "Ctrl+C should not type a 'c'",
            "Ctrl+C で 'c' を入力してはいけません"
        )
    );
}

#[test]
fn server_messages_update_state() {
    let mut app = App::new();
    app.on_server(text("Square\nA square."));
    app.on_server(ServerMsg::Map { map: sample_map() });
    app.on_server(ServerMsg::Status {
        status: sample_status(),
    });
    app.on_server(ServerMsg::Unknown);
    assert!(
        app.log
            .ends_with(&["Square".to_string(), "A square.".to_string()]),
        "{} {:?}",
        l(
            "text med \\n ska bli flera rader i loggen, loggen är",
            "text with \\n should become several log lines, the log is",
            "\\n を含むテキストはログの複数行になるべきです。ログ:"
        ),
        app.log
    );
    assert_eq!(app.map, Some(sample_map()));
    assert_eq!(app.status, Some(sample_status()));
}

#[test]
fn log_is_capped() {
    let mut app = App::new();
    for i in 0..MAX_LOG + 50 {
        app.on_server(text(&format!("line {i}")));
    }
    assert_eq!(app.log.len(), MAX_LOG);
    assert_eq!(app.log.last().unwrap(), &format!("line {}", MAX_LOG + 49));
}

#[test]
fn draws_status_log_map_and_input() {
    let mut app = App::new();
    app.on_server(text("Welcome, Anna!"));
    app.on_server(ServerMsg::Map { map: sample_map() });
    app.on_server(ServerMsg::Status {
        status: sample_status(),
    });
    type_text(&mut app, "look");

    let buf = render(&app, 80, 20);
    let s = screen(&buf);
    for want in ["Anna", "Square", "7/10", "Welcome, Anna!", "look"] {
        assert!(
            s.contains(want),
            "{} {want:?}:\n{s}",
            l(
                "skärmen ska visa",
                "the screen should show",
                "画面に表示されるべき:"
            )
        );
    }
    let forest = tile("forest").glyph.to_string();
    assert!(
        s.contains(&forest),
        "{} ({forest}):\n{s}",
        l(
            "kartan ska ritas med tile-tecken",
            "the map should be drawn with tile glyphs",
            "マップはタイル文字で描かれるべきです"
        )
    );

    // '@' should have the same colour as tile("self").
    let at = buf.content().iter().find(|c| c.symbol() == "@").expect(l(
        "spelaren '@' ska synas på kartan",
        "the player '@' should be visible on the map",
        "プレイヤー '@' がマップに表示されるべきです",
    ));
    let want = tile("self").fg;
    assert_eq!(
        at.fg,
        Color::Rgb(want.0, want.1, want.2),
        "{}",
        l(
            "'@' ska ritas i rätt färg",
            "'@' should be drawn in the right colour",
            "'@' は正しい色で描かれるべきです"
        )
    );
}

#[test]
fn log_scrolls_to_the_newest_lines() {
    let mut app = App::new();
    for i in 0..100 {
        app.on_server(text(&format!("message number {i}")));
    }
    let s = screen(&render(&app, 80, 20));
    assert!(
        s.contains("message number 99"),
        "{}:\n{s}",
        l(
            "senaste raden ska synas",
            "the newest line should be visible",
            "最新の行が表示されるべきです"
        )
    );
    assert!(
        !s.contains("message number 0 "),
        "{}:\n{s}",
        l(
            "äldsta raden ska ha scrollat bort",
            "the oldest line should have scrolled away",
            "最も古い行はスクロールで消えているべきです"
        )
    );
}

#[test]
fn rgb_converts_to_ratatui_color() {
    let c: Color = mud_client::render::Rgb(1, 2, 3).into();
    assert_eq!(c, Color::Rgb(1, 2, 3));
}
