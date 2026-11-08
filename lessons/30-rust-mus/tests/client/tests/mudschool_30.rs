// This file belongs to the course tool and is overwritten by `mudschool check`.

mod mudschool_common;

use std::time::{Duration, Instant};

use ratatui::crossterm::event::{KeyCode, KeyEvent, KeyModifiers};

use mud_client::app::{App, WALK_STEP};
use mud_client::messages;
use mud_client::mouse::{context_menu, describe_cell, find_path, screen_to_cell, step_command, WALK};
use mud_client::protocol::parse_line;
use mud_client::scene::{build_scene_at, draw_scene, TextRenderer};
use mudschool_common::l;

/// A 9×6 zone with a wall in the middle; Anna at (1,1), a rat at (6,1), a lamp at (6,4).
fn app() -> App {
    let zone = serde_json::json!({"type": "zone", "zone": {
        "room": "r", "width": 9, "height": 6,
        "rows": ["#########", "#...#...#", "#...#...#", "#.......#", "#....  ~#", "#########"],
        "legend": {"#": "wall", ".": "floor", "~": "water"},
        "entities": [
            {"id": "p:Anna", "kind": "player", "name": "Anna", "glyph": "@", "x": 1, "y": 1},
            {"id": "n:rat", "kind": "creature", "name": "a big rat", "glyph": "r", "x": 6, "y": 1},
            {"id": "i:lamp", "kind": "item", "name": "a lamp", "glyph": "*", "x": 6, "y": 3}
        ],
        "self": "p:Anna"}});
    let mut app = App::new();
    app.on_server(parse_line(&zone.to_string()));
    app
}

#[test]
fn a_star_in_the_client() {
    let app = app();
    let z = app.zone.as_ref().unwrap();
    let path = find_path(z, (1, 1), (7, 1)).unwrap_or_else(|| panic!("{}", l("en väg runt väggen", "a path round the wall", "壁を回る経路")));
    assert_eq!(path.len(), 6, "{}: {path:?}", l("den kortaste vägen med diagonala steg", "the shortest path with diagonal steps", "斜め移動を含む最短経路"));
    assert_eq!(*path.last().unwrap(), (7, 1));
    assert!(path.iter().all(|&(x, _)| x != 4 || path.contains(&(4, 3))), "{}", l("vägen går genom öppningen", "the path goes through the opening", "経路は開口部を通る"));
    assert_eq!(find_path(z, (1, 1), (4, 1)), None, "{}", l("in i en vägg: ingen väg", "into a wall: no path", "壁の中へ: 経路なし"));
    assert_eq!(find_path(z, (1, 1), (5, 4)), None, "{}", l("till en okänd ruta: ingen väg", "to an unknown cell: no path", "未知のマスへ: 経路なし"));
    assert_eq!(find_path(z, (1, 1), (1, 1)), Some(vec![]));
}

#[test]
fn steps_have_names() {
    assert_eq!(step_command((3, 3), (3, 2)), Some("north"));
    assert_eq!(step_command((3, 3), (4, 4)), Some("southeast"));
    assert_eq!(step_command((3, 3), (2, 2)), Some("northwest"));
    assert_eq!(step_command((3, 3), (5, 3)), None);
}

#[test]
fn the_tooltip() {
    let app = app();
    let z = app.zone.as_ref().unwrap();
    assert_eq!(describe_cell(z, 6, 1).as_deref(), Some("a big rat"));
    assert_eq!(describe_cell(z, 4, 1).as_deref(), Some(messages::tile_name("wall")));
    assert_eq!(describe_cell(z, 5, 4), None, "{}", l("okända rutor har inget namn", "unknown cells have no name", "未知のマスには名前がない"));
}

#[test]
fn the_right_click_menu() {
    let app = app();
    let z = app.zone.as_ref().unwrap();
    let commands = |x, y| context_menu(z, x, y).into_iter().map(|m| m.command).collect::<Vec<_>>();
    assert_eq!(commands(6, 1), vec!["examine rat", "kill rat", WALK], "{}", l(
        "en varelse: titta, anfall (med ID:t utan prefix), gå hit",
        "a creature: look, attack (with the ID without its prefix), walk here",
        "生き物: 調べる、攻撃する（接頭辞を除いた ID で）、ここへ歩く"
    ));
    assert_eq!(commands(6, 3), vec!["examine lamp", "get lamp", WALK]);
    assert_eq!(commands(2, 2), vec![WALK]);
    assert!(commands(4, 1).is_empty(), "{}", l("en vägg har ingen meny", "a wall has no menu", "壁にはメニューがない"));
    let labels = context_menu(z, 6, 1);
    assert!(labels[1].label.contains("a big rat") && labels[1].label.starts_with(messages::ATTACK));
}

#[test]
fn click_to_walk() {
    let mut app = app();
    let now = Instant::now();
    assert_eq!(app.left_click((3, 3), None), None, "{}", l("att gå skickar inget direkt", "walking sends nothing at once", "歩く操作はすぐには何も送らない"));
    assert_eq!(app.next_walk_step(now).as_deref(), Some("southeast"));
    assert_eq!(app.next_walk_step(now + Duration::from_millis(10)), None, "{}", l("nästa steg först efter WALK_STEP", "the next step only after WALK_STEP", "次の 1 歩は WALK_STEP の後"));
    assert_eq!(app.next_walk_step(now + WALK_STEP).as_deref(), Some("southeast"));
    assert_eq!(app.next_walk_step(now + WALK_STEP * 2), None, "{}", l("framme", "arrived", "到着"));

    app.left_click((7, 1), None);
    assert!(!app.walk.is_empty());
    app.on_key(KeyEvent::new(KeyCode::Char('x'), KeyModifiers::NONE));
    assert!(app.walk.is_empty(), "{}", l("en tangent avbryter promenaden", "a key stops the walk", "キーを押すと歩行が止まる"));
}

#[test]
fn choosing_in_the_menu() {
    let mut app = app();
    app.right_click((6, 1));
    assert_eq!(app.menu.as_ref().unwrap().items.len(), 3);
    assert_eq!(app.left_click((6, 2), Some(1)).as_deref(), Some("kill rat"), "{}", l("rad 1 i menyn: anfall", "row 1 of the menu: attack", "メニューの 1 行目: 攻撃"));
    assert!(app.menu.is_none(), "{}", l("menyn stängs efter ett val", "the menu closes after a choice", "選んだらメニューは閉じる"));
    app.right_click((2, 2));
    assert_eq!(app.left_click((2, 2), Some(0)), None);
    assert!(!app.walk.is_empty(), "{}", l("\"Gå hit\" startar en promenad", "\"Walk here\" starts a walk", "「ここへ歩く」で歩き始める"));
    app.right_click((4, 1));
    assert!(app.menu.is_none(), "{}", l("ingen meny för en vägg", "no menu for a wall", "壁にはメニューを出さない"));
}

#[test]
fn the_scene_maps_the_mouse() {
    let mut app = app();
    app.hover = Some((6, 1));
    let scene = build_scene_at(&app, 20, 14, Instant::now());
    assert_eq!(scene.zone_cell(6, 2), Some((6, 1)), "{}", l("skärmrad 1 är zonens rad 0", "screen row 1 is the zone's row 0", "画面の 1 行目がゾーンの 0 行目"));
    assert_eq!(scene.zone_cell(6, 0), None, "{}", l("statusraden är inte zonen", "the status line is not the zone", "ステータス行はゾーンではない"));
    assert_eq!(screen_to_cell(70.0, 35.0, 32.0), (2, 1));
    let mut text = TextRenderer::new(20, 14);
    draw_scene(&scene, &mut text);
    assert!(text.lines().iter().any(|l| l.contains("a big rat")), "{}\n{}", l(
        "verktygstipset ska ritas (overlay)",
        "the tooltip should be drawn (overlay)",
        "ツールチップが描かれるべきです（overlay）"
    ), text.lines().join("\n"));
}

#[test]
fn a_new_room_stops_the_walk() {
    let mut app = app();
    app.left_click((3, 3), None);
    app.right_click((6, 1));
    assert!(!app.walk.is_empty() && app.menu.is_some());
    let other = serde_json::json!({"type": "zone", "zone": {
        "room": "elsewhere", "width": 3, "height": 3, "rows": ["...", "...", "..."],
        "legend": {".": "floor"},
        "entities": [{"id": "p:Anna", "kind": "player", "name": "Anna", "glyph": "@", "x": 1, "y": 1}],
        "self": "p:Anna"}});
    app.on_server(parse_line(&other.to_string()));
    assert!(app.walk.is_empty() && app.menu.is_none(), "{}", l(
        "i ett nytt rum hör promenaden och menyn till det gamla – de ska försvinna",
        "in a new room the walk and the menu belong to the old one – they should go",
        "新しい部屋では、歩行とメニューは古い部屋のものなので消すべきです"
    ));
}

#[test]
fn the_menu_knows_where_it_is() {
    let mut app = app();
    app.right_click((6, 1));
    let menu = app.menu.clone().unwrap();
    // Drawn one column right of the cell, one row per entry: screen (7, 2), (7, 3), (7, 4).
    assert_eq!(menu.row_at((0, 0), 7, 2), Some(0));
    assert_eq!(menu.row_at((0, 0), 8, 4), Some(2));
    assert_eq!(menu.row_at((0, 0), 6, 2), None, "{}", l("till vänster om menyn", "left of the menu", "メニューの左"));
    assert_eq!(menu.row_at((0, 0), 7, 5), None, "{}", l("under menyn", "below the menu", "メニューの下"));
    assert_eq!(menu.row_at((0, 0), 60, 2), None, "{}", l("långt till höger", "far to the right", "はるか右"));
}
