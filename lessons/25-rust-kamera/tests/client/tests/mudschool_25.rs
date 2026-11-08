// This file belongs to the course tool and is overwritten by `mudschool check`.

mod mudschool_common;

use ratatui::backend::TestBackend;
use ratatui::crossterm::event::{KeyCode, KeyEvent, KeyEventKind, KeyEventState, KeyModifiers};
use ratatui::Terminal;

use mud_client::app::{draw, App};
use mud_client::protocol::{parse_line, ServerMsg};
use mud_client::zone::{camera, ZoneView};
use mudschool_common::l;

fn key(code: KeyCode) -> KeyEvent {
    KeyEvent {
        code,
        modifiers: KeyModifiers::NONE,
        kind: KeyEventKind::Press,
        state: KeyEventState::NONE,
    }
}

/// A 40×12 zone: walls around, a pond, the player at (20,6), a cat and a lamp.
fn zone_json() -> String {
    let mut rows = vec!["#".repeat(40)];
    for y in 1..11 {
        let mut row: Vec<char> = format!("#{}#", ".".repeat(38)).chars().collect();
        if y == 3 {
            row[30] = '~';
        }
        rows.push(row.into_iter().collect());
    }
    rows.push("#".repeat(40));
    let rows: Vec<String> = rows.iter().map(|r| format!("\"{r}\"")).collect();
    format!(
        r##"{{"type":"zone","zone":{{"room":"torg","width":40,"height":12,"rows":[{}],
        "legend":{{"#":"wall",".":"floor","~":"water"}},
        "entities":[{{"id":"p:Anna","kind":"player","name":"Anna","glyph":"@","x":20,"y":6}},
                    {{"id":"n:cat","kind":"creature","name":"a cat","glyph":"c","x":22,"y":6}},
                    {{"id":"i:lamp","kind":"item","name":"a lamp","glyph":"*","x":5,"y":2}}],
        "self":"p:Anna"}}}}"##,
        rows.join(",")
    )
    .replace('\n', "")
}

fn screen(app: &App, w: u16, h: u16) -> Vec<String> {
    let mut terminal = Terminal::new(TestBackend::new(w, h)).unwrap();
    terminal.draw(|f| draw(f, app)).unwrap();
    let buf = terminal.backend().buffer();
    (0..h)
        .map(|y| (0..w).map(|x| buf[(x, y)].symbol()).collect::<String>())
        .collect()
}

#[test]
fn the_camera() {
    // (focus, view, size) -> first cell
    let cases = [
        (20, 10, 40, 15), // in the middle: focus - view/2
        (2, 10, 40, 0),   // near the start: clamped to 0
        (38, 10, 40, 30), // near the end: clamped to size - view
        (5, 50, 40, 0),   // the zone fits: 0
        (0, 1, 1, 0),
    ];
    for (focus, view, size, want) in cases {
        assert_eq!(
            camera(focus, view, size),
            want,
            "camera({focus}, {view}, {size}) – {}",
            l(
                "fokus i mitten, men aldrig utanför zonen",
                "focus in the middle, but never outside the zone",
                "注目点を中央に、ただしゾーンの外は見せない"
            )
        );
    }
}

#[test]
fn parses_zone_and_delta() {
    let ServerMsg::Zone { zone } = parse_line(&zone_json()) else {
        panic!("{}", l("en zone-frame ska bli ServerMsg::Zone", "a zone frame should become ServerMsg::Zone", "zone フレームは ServerMsg::Zone になるべきです"));
    };
    assert_eq!((zone.width, zone.height, zone.rows.len()), (40, 12, 12));
    assert_eq!(zone.self_id, "p:Anna", "\"self\" -> self_id");
    assert_eq!(zone.entities.len(), 3);

    let delta = parse_line(r#"{"type":"delta","delta":{"room":"torg","moved":[{"id":"p:Anna","kind":"player","name":"Anna","glyph":"@","x":21,"y":6}],"removed":["i:lamp"]}}"#);
    let ServerMsg::Delta { delta } = delta else {
        panic!("{}", l("en delta-frame ska bli ServerMsg::Delta", "a delta frame should become ServerMsg::Delta", "delta フレームは ServerMsg::Delta になるべきです"));
    };
    assert!(delta.added.is_empty(), "added: #[serde(default)]");
    assert_eq!(delta.removed, vec!["i:lamp".to_string()]);
}

#[test]
fn deltas_update_the_view() {
    let ServerMsg::Zone { zone } = parse_line(&zone_json()) else { unreachable!() };
    let mut view = ZoneView::from(zone);
    assert_eq!(view.self_pos(), Some((20, 6)));
    assert_eq!(view.meaning_at(30, 3), Some("water"));
    assert_eq!(view.code_at(99, 99), ' ', "{}", l("utanför zonen: ' '", "outside the zone: ' '", "ゾーンの外: ' '"));

    let ServerMsg::Delta { delta } = parse_line(r#"{"type":"delta","delta":{"room":"torg","added":[{"id":"p:Bo","kind":"player","glyph":"@","x":1,"y":1}],"moved":[{"id":"p:Anna","kind":"player","glyph":"@","x":21,"y":7}],"removed":["n:cat"]}}"#) else { unreachable!() };
    assert!(view.apply(delta));
    assert_eq!(view.self_pos(), Some((21, 7)));
    assert!(view.entities.contains_key("p:Bo"));
    assert!(!view.entities.contains_key("n:cat"));

    let ServerMsg::Delta { delta } = parse_line(r#"{"type":"delta","delta":{"room":"hall","removed":["p:Bo"]}}"#) else { unreachable!() };
    assert!(
        !view.apply(delta) && view.entities.contains_key("p:Bo"),
        "{}",
        l(
            "en delta för ett annat rum ska ignoreras",
            "a delta for another room should be ignored",
            "別の部屋の delta は無視するべきです"
        )
    );
}

#[test]
fn the_zone_is_drawn_around_the_player() {
    let mut app = App::new();
    app.on_server(parse_line(&zone_json()));
    let lines = screen(&app, 60, 24).join("\n");
    assert!(lines.contains("@.c"), "{}\n{lines}", l(
        "zonen ska ritas med Anna (@) och katten (c) två steg österut",
        "the zone should be drawn with Anna (@) and the cat (c) two steps east",
        "ゾーンにはアンナ（@）と、2 マス東の猫（c）が描かれるべきです"
    ));
    // Walk to the far east edge: the camera follows but never shows outside.
    let ServerMsg::Delta { delta } = parse_line(r#"{"type":"delta","delta":{"room":"torg","moved":[{"id":"p:Anna","kind":"player","glyph":"@","x":38,"y":6}]}}"#) else { unreachable!() };
    app.zone.as_mut().unwrap().apply(delta);
    let lines = screen(&app, 60, 24);
    let row = lines.iter().find(|l| l.contains('@')).expect("@");
    assert!(row.contains(".@#"), "{}\n{}", l(
        "vid östra kanten ska väggen synas direkt till höger om @",
        "at the east edge the wall should be right next to the @",
        "東の端では @ のすぐ右に壁が見えるべきです"
    ), lines.join("\n"));
}

#[test]
fn arrows_walk_when_the_input_is_empty() {
    let mut app = App::new();
    assert_eq!(app.on_key(key(KeyCode::Up)), Some("north".to_string()));
    assert_eq!(app.on_key(key(KeyCode::Left)), Some("west".to_string()));
    assert_eq!(app.on_key(key(KeyCode::Right)), Some("east".to_string()));
    assert_eq!(app.on_key(key(KeyCode::Down)), Some("south".to_string()));
    assert!(app.log.is_empty() && app.history.is_empty(), "{}", l(
        "promenader ska inte hamna i loggen eller historiken",
        "walking should not end up in the log or the history",
        "歩く操作はログや履歴に残すべきではありません"
    ));

    // While typing, Up is the history again.
    for c in "look".chars() {
        app.on_key(key(KeyCode::Char(c)));
    }
    app.on_key(key(KeyCode::Enter));
    app.on_key(key(KeyCode::Char('x')));
    assert_eq!(app.on_key(key(KeyCode::Up)), None);
    assert_eq!(app.input, "look", "{}", l(
        "när man skriver ska Upp bläddra i historiken",
        "while typing, Up should browse the history",
        "入力中は上キーで履歴をたどるべきです"
    ));
}
