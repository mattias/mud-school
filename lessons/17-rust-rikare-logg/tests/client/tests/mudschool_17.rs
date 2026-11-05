// This file belongs to the course tool and is overwritten by `mudschool check`.

mod mudschool_common;

use ratatui::backend::TestBackend;
use ratatui::buffer::Buffer;
use ratatui::crossterm::event::{KeyCode, KeyEvent, KeyModifiers};
use ratatui::Terminal;

use mud_client::app::{draw, style_for, App, LogLine, PAGE};
use mud_client::protocol::{parse_line, Kind, ServerMsg};
use mudschool_common::l;

fn key(code: KeyCode) -> KeyEvent {
    KeyEvent::new(code, KeyModifiers::NONE)
}

fn type_line(app: &mut App, text: &str) -> Option<String> {
    for c in text.chars() {
        app.on_key(key(KeyCode::Char(c)));
    }
    app.on_key(key(KeyCode::Enter))
}

fn render(app: &App) -> Buffer {
    let mut terminal = Terminal::new(TestBackend::new(80, 20)).unwrap();
    terminal.draw(|f| draw(f, app)).unwrap();
    terminal.backend().buffer().clone()
}

#[test]
fn parses_kinds() {
    let cases = [
        (r#"{"type":"text","text":"hi","kind":"say"}"#, Kind::Say),
        (r#"{"type":"text","text":"hi","kind":"room"}"#, Kind::Room),
        (r#"{"type":"text","text":"hi","kind":"move"}"#, Kind::Move),
        (r#"{"type":"text","text":"hi","kind":"error"}"#, Kind::Error),
        (
            r#"{"type":"text","text":"hi","kind":"combat"}"#,
            Kind::Combat,
        ),
        (r#"{"type":"text","text":"hi","kind":"info"}"#, Kind::Info),
        (r#"{"type":"text","text":"hi"}"#, Kind::Info),
        (
            r#"{"type":"text","text":"hi","kind":"poetry"}"#,
            Kind::Other,
        ),
    ];
    for (line, want) in cases {
        match parse_line(line) {
            ServerMsg::Text { kind, .. } => assert_eq!(kind, want, "{line}"),
            other => panic!("{line}: {other:?}"),
        }
    }
    assert_eq!(
        parse_line("plain text"),
        ServerMsg::Text {
            text: "plain text".into(),
            kind: Kind::Info
        }
    );
}

#[test]
fn log_keeps_kinds() {
    let mut app = App::new();
    app.on_server(parse_line(
        r#"{"type":"text","text":"Square\nA square.","kind":"room"}"#,
    ));
    assert_eq!(
        app.log[app.log.len() - 2..],
        [
            LogLine {
                kind: Kind::Room,
                text: "Square".into()
            },
            LogLine {
                kind: Kind::Room,
                text: "A square.".into()
            }
        ]
    );
}

#[test]
fn kinds_have_different_styles() {
    let kinds = [Kind::Room, Kind::Say, Kind::Move, Kind::Error, Kind::Combat];
    for (i, a) in kinds.iter().enumerate() {
        for b in &kinds[i + 1..] {
            assert_ne!(
                style_for(*a),
                style_for(*b),
                "{} {a:?} / {b:?}",
                l(
                    "olika kinds ska ritas olika:",
                    "different kinds should be drawn differently:",
                    "種類が違えば描き方も変えるべきです:"
                )
            );
        }
    }
    assert_eq!(
        style_for(Kind::Other),
        style_for(Kind::Info),
        "{}",
        l(
            "okända kinds ska se ut som Info",
            "unknown kinds should look like Info",
            "未知の種類は Info と同じ見た目にするべきです"
        )
    );
}

#[test]
fn log_lines_are_drawn_in_their_style() {
    let mut app = App::new();
    app.on_server(parse_line(
        r#"{"type":"text","text":"Bertil says: hello","kind":"say"}"#,
    ));
    let buf = render(&app);
    let area = buf.area;
    let mut found = None;
    for y in 0..area.height {
        let row: String = (0..area.width).map(|x| buf[(x, y)].symbol()).collect();
        if let Some(col) = row.find("Bertil says") {
            let x = row[..col].chars().count() as u16;
            found = Some(buf[(x, y)].fg);
        }
    }
    let fg = found.expect(l(
        "raden ska synas i loggen",
        "the line should be visible in the log",
        "行がログに表示されるべきです",
    ));
    assert_eq!(
        Some(fg),
        style_for(Kind::Say).fg,
        "{}",
        l(
            "say-raden ska ritas med style_for(Kind::Say)",
            "the say line should be drawn with style_for(Kind::Say)",
            "say の行は style_for(Kind::Say) で描くべきです"
        )
    );
}

#[test]
fn history_with_up_and_down() {
    let mut app = App::new();
    assert_eq!(type_line(&mut app, "look").as_deref(), Some("look"));
    type_line(&mut app, "north");
    type_line(&mut app, "north"); // the same twice in a row is stored once
    type_line(&mut app, "say hi");
    assert_eq!(app.history, ["look", "north", "say hi"]);

    for c in "dra".chars() {
        app.on_key(key(KeyCode::Char(c)));
    }
    app.on_key(key(KeyCode::Up));
    assert_eq!(app.input, "say hi");
    app.on_key(key(KeyCode::Up));
    assert_eq!(app.input, "north");
    app.on_key(key(KeyCode::Up));
    app.on_key(key(KeyCode::Up)); // stays at the oldest
    assert_eq!(app.input, "look");
    app.on_key(key(KeyCode::Down));
    assert_eq!(app.input, "north");
    app.on_key(key(KeyCode::Down));
    app.on_key(key(KeyCode::Down));
    assert_eq!(
        app.input,
        "dra",
        "{}",
        l(
            "efter sista Down ska det man höll på att skriva komma tillbaka",
            "after the last Down, what you were typing should come back",
            "最後の Down の後は、入力途中だった内容が戻るべきです"
        )
    );
    // Using a history entry sends it.
    app.on_key(key(KeyCode::Backspace));
    app.on_key(key(KeyCode::Backspace));
    app.on_key(key(KeyCode::Backspace));
    app.on_key(key(KeyCode::Up));
    assert_eq!(app.on_key(key(KeyCode::Enter)).as_deref(), Some("say hi"));
}

#[test]
fn up_with_empty_history_does_nothing() {
    let mut app = App::new();
    app.on_key(key(KeyCode::Up));
    app.on_key(key(KeyCode::Down));
    assert_eq!(app.input, "");
}

#[test]
fn page_up_and_down_scroll_the_log() {
    let mut app = App::new();
    for i in 0..100 {
        app.on_server(ServerMsg::Text {
            text: format!("line {i}"),
            kind: Kind::Info,
        });
    }
    let screen = |app: &App| -> String {
        let buf = render(app);
        let a = buf.area;
        (0..a.height)
            .map(|y| {
                (0..a.width)
                    .map(|x| buf[(x, y)].symbol())
                    .collect::<String>()
            })
            .collect::<Vec<_>>()
            .join("\n")
    };
    assert!(screen(&app).contains("line 99"));

    app.on_key(key(KeyCode::PageUp));
    assert_eq!(app.scroll, PAGE);
    let s = screen(&app);
    assert!(
        !s.contains("line 99") && s.contains(&format!("line {}", 99 - PAGE)),
        "{}\n{s}",
        l(
            "efter PageUp ska loggen visa äldre rader",
            "after PageUp the log should show older lines",
            "PageUp の後、ログは古い行を表示するべきです"
        )
    );

    // New messages don't move the view while scrolled up.
    app.on_server(ServerMsg::Text {
        text: "new line".into(),
        kind: Kind::Info,
    });
    assert!(!screen(&app).contains("new line"));
    assert_eq!(
        app.scroll,
        PAGE + 1,
        "{}",
        l(
            "när loggen är uppskrollad ska en ny rad öka scroll med 1, så att vyn står still",
            "while the log is scrolled up, a new line should increase scroll by 1, so that the view stays still",
            "ログを上にスクロールしている間は、新しい行で scroll を 1 増やし、表示を動かさないようにするべきです"
        )
    );
    assert!(screen(&app).contains(&format!("line {}", 99 - PAGE)));

    app.on_key(key(KeyCode::PageDown));
    app.on_key(key(KeyCode::PageDown));
    assert_eq!(app.scroll, 0);
    assert!(screen(&app).contains("new line"));

    for _ in 0..50 {
        app.on_key(key(KeyCode::PageUp));
    }
    assert!(app.scroll <= app.log.len());
    app.on_key(key(KeyCode::End));
    assert_eq!(app.scroll, 0);
}
