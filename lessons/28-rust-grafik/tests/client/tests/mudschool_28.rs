// This file belongs to the course tool and is overwritten by `mudschool check`.

mod mudschool_common;

use mud_client::app::App;
use mud_client::protocol::parse_line;
use mud_client::render::Rgb;
use mud_client::scene::{build_scene, draw_scene, shade, zone_rows, Renderer, SceneCell, TextRenderer, FOG, FULL_LIGHT};
use mud_client::tileset::{atlas_rect, entity_sprite, tile_sprite, TILE_PX};
use mudschool_common::l;

/// A 9×5 zone: Anna at (4,2), a cat at (6,2), the left part not visible.
fn zone_json(rows: [&str; 5]) -> String {
    serde_json::json!({"type": "zone", "zone": {
        "room": "torg", "width": 9, "height": 5,
        "rows": rows,
        "legend": {"#": "wall", ".": "floor", "~": "water"},
        "entities": [
            {"id": "p:Anna", "kind": "player", "name": "Anna", "glyph": "@", "x": 4, "y": 2},
            {"id": "n:cat", "kind": "creature", "name": "a cat", "glyph": "c", "x": 6, "y": 2}
        ],
        "self": "p:Anna",
        "visible": ["000111111", "000111111", "000111111", "000111111", "000111111"]
    }})
    .to_string()
}

const ROWS: [&str; 5] = ["#########", "#.......#", "#..~....#", "#.......#", "####.####"];

fn app() -> App {
    let mut app = App::new();
    app.on_server(parse_line(&zone_json(ROWS)));
    app
}

#[test]
fn the_tileset() {
    assert_eq!(tile_sprite("floor"), Some(0));
    assert_eq!(tile_sprite("wall"), Some(1));
    assert_eq!(tile_sprite("water"), Some(4));
    assert_eq!(tile_sprite("stairs_down"), Some(8));
    assert_eq!(tile_sprite("lava"), None, "{}", l("okända betydelser har ingen bild", "unknown meanings have no picture", "未知の意味には絵がありません"));
    assert_eq!(entity_sprite("player", true), 9);
    assert_eq!(entity_sprite("player", false), 10);
    assert_eq!(entity_sprite("creature", false), 11);
    assert_eq!(entity_sprite("item", false), 12);
    assert_eq!(TILE_PX, 16.0);
    assert_eq!(atlas_rect(0), (0.0, 0.0, 16.0, 16.0));
    assert_eq!(atlas_rect(4), (64.0, 0.0, 16.0, 16.0));
    assert_eq!(atlas_rect(17), (16.0, 16.0, 16.0, 16.0), "{}", l(
        "16 bilder per rad i atlasen: index 17 är rad 1, kolumn 1",
        "16 pictures per row in the atlas: index 17 is row 1, column 1",
        "アトラスは 1 行に 16 枚: index 17 は 1 行目の 1 列目"
    ));
}

#[test]
fn shading() {
    assert_eq!(shade(Rgb(200, 100, 50), FULL_LIGHT), Rgb(200, 100, 50));
    assert_eq!(shade(Rgb(200, 100, 50), Rgb(0, 0, 0)), Rgb(0, 0, 0));
    assert_eq!(shade(Rgb(255, 255, 255), Rgb(128, 64, 255)), Rgb(128, 64, 255));
}

#[test]
fn deltas_reveal_tiles_and_change_visibility() {
    let mut app = App::new();
    app.on_server(parse_line(&zone_json(["#########", "#.......#", "   ~....#", "#.......#", "####.####"])));
    let z = app.zone.as_ref().unwrap();
    assert_eq!(z.code_at(1, 2), ' ');
    assert!(!z.is_visible(1, 2) && z.is_visible(5, 2));
    app.on_server(parse_line(r#"{"type":"delta","delta":{"room":"torg","tiles":[{"x":1,"y":2,"code":"T","meaning":"tree"}],"visible":["111111111","111111111","111111111","111111111","111111111"]}}"#));
    let z = app.zone.as_ref().unwrap();
    assert_eq!(z.code_at(1, 2), 'T', "{}", l("deltans tiles ska läggas in i zonen", "the delta's tiles should be put into the zone", "delta の tiles をゾーンに入れるべきです"));
    assert_eq!(z.meaning_at(1, 2), Some("tree"), "{}", l("och deras betydelse i förklaringen", "and their meaning into the legend", "その意味も凡例に入れるべきです"));
    assert!(z.is_visible(1, 2), "{}", l("deltans visible ersätter synfältet", "the delta's visible replaces the field of view", "delta の visible が視界を置き換えるべきです"));
}

#[test]
fn the_scene() {
    let scene = build_scene(&app(), 20, 14);
    assert_eq!(zone_rows(14), 14 - 2 - 5, "{}", l("zonens höjd: allt utom status, 5 loggrader och inmatningen", "the zone's height: everything except the status, 5 log lines and the input", "ゾーンの高さ: ステータス、ログ 5 行、入力欄以外のすべて"));
    let at = |x: u16, y: u16| scene.cells.iter().find(|c| c.x == x && c.y == y).cloned();
    let anna = at(4, 2).expect("(4,2)");
    assert_eq!((anna.tile, anna.entity, anna.glyph), (Some(0), Some(9), '@'), "{}", l(
        "Anna: golvet under (0), hennes bild ovanpå (9), tecknet @",
        "Anna: the floor below (0), her picture on top (9), the glyph @",
        "アンナ: 下に床（0）、上に自分の絵（9）、文字 @"
    ));
    assert_eq!(at(6, 2).unwrap().entity, Some(11));
    assert_eq!(at(3, 2).unwrap().tile, Some(4), "~ = water = 4");
    assert_eq!(anna.light, FULL_LIGHT);
    assert_eq!(at(1, 1).unwrap().light, FOG, "{}", l(
        "utforskat men inte synligt: dimma (FOG)",
        "explored but not visible: fog (FOG)",
        "探索済みだが見えない: 霧（FOG）"
    ));
    assert!(scene.status.contains(mud_client::messages::NOT_LOGGED_IN.trim()));
    assert_eq!(scene.input, "> ");
}

/// A renderer that only counts – any type can implement the trait.
#[derive(Default)]
struct Counter {
    cells: usize,
    texts: Vec<(u16, u16, String)>,
}

impl Renderer for Counter {
    fn cell(&mut self, _x: u16, _y: u16, _cell: &SceneCell) {
        self.cells += 1;
    }
    fn text(&mut self, x: u16, y: u16, text: &str, _fg: Rgb) {
        self.texts.push((x, y, text.to_string()));
    }
}

#[test]
fn any_renderer_can_draw_a_scene() {
    let mut app = app();
    app.on_server(parse_line(r#"{"type":"text","text":"Hello"}"#));
    let scene = build_scene(&app, 20, 14);

    let mut counter = Counter::default();
    draw_scene(&scene, &mut counter);
    assert_eq!(counter.cells, scene.cells.len());
    assert!(counter.texts.iter().any(|(_, y, t)| *y == 13 && t == "> "), "{}", l("inmatningen på sista raden", "the input on the last line", "入力欄は最後の行"));

    let mut text = TextRenderer::new(20, 14);
    draw_scene(&scene, &mut text);
    let lines = text.lines();
    let row: String = lines[3].chars().take(9).collect();
    assert_eq!(row, "#..≈@.c.#", "{}\n{}", l(
        "TextRenderer: zonen börjar på rad 1 (efter statusraden)",
        "TextRenderer: the zone starts on row 1 (after the status line)",
        "TextRenderer: ゾーンは 1 行目（ステータス行の次）から"
    ), lines.join("\n"));
    assert!(lines.iter().any(|l| l.starts_with("Hello")));
}
