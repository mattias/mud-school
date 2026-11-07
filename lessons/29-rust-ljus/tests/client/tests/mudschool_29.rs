// This file belongs to the course tool and is overwritten by `mudschool check`.

mod mudschool_common;

use std::time::{Duration, Instant};

use mud_client::app::App;
use mud_client::light::{ambient, light_at, lights_for, shake_offset, shimmer, Light, Particles, TORCH};
use mud_client::protocol::parse_line;
use mud_client::render::Rgb;
use mud_client::scene::{build_scene_at, Renderer, SceneCell, FOG};
use mudschool_common::l;

fn cave_app() -> App {
    let mut app = App::new();
    let mut rows = vec!["#".repeat(21)];
    for y in 1..8 {
        let mut row: Vec<char> = format!("#{}#", ".".repeat(19)).chars().collect();
        if y == 4 {
            row[15] = '~';
        }
        rows.push(row.into_iter().collect());
    }
    rows.push("#".repeat(21));
    let zone = serde_json::json!({"type": "zone", "zone": {
        "room": "g", "width": 21, "height": 9, "rows": rows,
        "legend": {"#": "wall", ".": "floor", "~": "water"},
        "entities": [{"id": "p:Anna", "kind": "player", "name": "Anna", "glyph": "@", "x": 5, "y": 4}],
        "self": "p:Anna"}});
    app.on_server(parse_line(&zone.to_string()));
    app.on_server(parse_line(r#"{"type":"status","status":{"name":"Anna","room":"Grotto","terrain":"cave","hp":20,"max_hp":20,"online":1}}"#));
    app
}

#[test]
fn ambient_light() {
    assert_eq!(ambient(""), Rgb(255, 255, 255), "{}", l("okänd terräng: fullt ljus", "unknown terrain: full light", "未知の地形: 最大の明るさ"));
    let cave = ambient("cave");
    let town = ambient("town");
    assert!(cave.0 < 60 && cave.1 < 60, "{}", l("en grotta är mörk", "a cave is dark", "洞窟は暗い"));
    assert!(town.0 > 150, "{}", l("en stad är ljus", "a town is light", "町は明るい"));
}

#[test]
fn lights_blend() {
    let dark = Rgb(0, 0, 0);
    let red = Light { x: 0.0, y: 0.0, radius: 4.0, color: Rgb(200, 0, 0) };
    let green = Light { x: 2.0, y: 0.0, radius: 4.0, color: Rgb(0, 200, 0) };
    assert_eq!(light_at(0, 0, &[red], dark), Rgb(200, 0, 0), "{}", l("mitt i ljuset: hela färgen", "in the middle of the light: the whole colour", "光の中心: 色そのもの"));
    assert_eq!(light_at(2, 0, &[red], dark), Rgb(100, 0, 0), "{}", l("halva radien bort: halva färgen", "half the radius away: half the colour", "半径の半分離れると: 色も半分"));
    assert_eq!(light_at(5, 0, &[red], dark), dark, "{}", l("utanför radien: inget", "outside the radius: nothing", "半径の外: なし"));
    assert_eq!(light_at(1, 0, &[red, green], dark), Rgb(150, 150, 0), "{}", l("rött och grönt blir gult", "red and green make yellow", "赤と緑で黄色"));
    assert_eq!(light_at(0, 0, &[red, red], Rgb(100, 0, 0)), Rgb(255, 0, 0), "{}", l("kanalerna slutar på 255", "the channels stop at 255", "各チャンネルは 255 で止まる"));
}

#[test]
fn every_player_carries_a_torch() {
    let app = cave_app();
    let lights = lights_for(app.zone.as_ref().unwrap());
    assert_eq!(lights.len(), 1);
    assert_eq!((lights[0].x, lights[0].y, lights[0].color), (5.0, 4.0, TORCH));
}

#[test]
fn the_cave_is_dark_except_near_the_torch() {
    let app = cave_app();
    let now = Instant::now();
    let scene = build_scene_at(&app, 21, 16, now);
    let light = |x: u16, y: u16| scene.cells.iter().find(|c| c.x == x && c.y == y).unwrap().light;
    let near = light(6, 4);
    let far = light(18, 1);
    assert!(near.0 > 150, "{}: {near:?}", l("bredvid facklan är det ljust", "next to the torch it is light", "松明の隣は明るい"));
    assert!(far.0 < 60, "{}: {far:?}", l("långt bort i grottan är det mörkt", "far away in the cave it is dark", "洞窟の奥は暗い"));
}

#[test]
fn water_shimmers() {
    assert_eq!(shimmer(3, 4, 1000), shimmer(3, 4, 1000));
    assert_ne!(shimmer(3, 4, 0), shimmer(3, 4, 300), "{}", l("vattnet ändras med tiden", "the water changes with time", "水は時間とともに変わる"));
    assert_ne!(shimmer(3, 4, 0), shimmer(5, 4, 0), "{}", l("och med platsen", "and with the place", "場所によっても変わる"));
    for t in (0..5000).step_by(37) {
        let s = shimmer(1, 1, t);
        assert!((0.7..=1.0).contains(&s), "shimmer = {s}");
    }
    let app = cave_app();
    let start = app.started.unwrap();
    let a = build_scene_at(&app, 21, 16, start);
    let b = build_scene_at(&app, 21, 16, start + Duration::from_millis(500));
    let water = |s: &mud_client::scene::Scene| s.cells.iter().find(|c| c.x == 15 && c.y == 4).unwrap().light;
    assert_ne!(water(&a), water(&b), "{}", l("vattnets ljus ska skifta mellan två tidpunkter", "the water's light should change between two moments", "水の光は 2 つの時刻で変わるべきです"));
}

#[test]
fn hits_make_sparks_that_fade() {
    let mut app = cave_app();
    let now = Instant::now();
    app.tick(now);
    app.on_server_at(parse_line(r#"{"type":"status","status":{"name":"Anna","room":"Grotto","terrain":"cave","hp":15,"max_hp":20,"online":1}}"#), now);
    assert!(app.particles.list.len() >= 8, "{}", l("en träff ska ge gnistor där du står", "a hit should give sparks where you stand", "攻撃を受けたら、自分の場所に火花が出るべきです"));
    let p = app.particles.list[0];
    assert!((p.x - 5.5).abs() < 0.01 && (p.y - 4.5).abs() < 0.01, "{}", l("gnistorna börjar mitt i din ruta", "the sparks start in the middle of your cell", "火花は自分のマスの中央から始まる"));
    app.tick(now + Duration::from_millis(100));
    let moved = app.particles.list[0];
    assert!(moved.x != p.x || moved.y != p.y, "{}", l("tick ska flytta gnistorna", "tick should move the sparks", "tick で火花が動くべきです"));
    assert!(moved.alpha() < 1.0);
    app.tick(now + Duration::from_secs(2));
    assert!(app.particles.list.is_empty(), "{}", l("döda gnistor ska försvinna", "dead sparks should disappear", "消えた火花は取り除くべきです"));

    let mut ps = Particles::default();
    ps.burst(0.0, 0.0, 6, Rgb(255, 0, 0));
    assert_eq!(ps.list.len(), 6);
}

#[test]
fn the_screen_shakes_while_it_flashes() {
    let now = Instant::now();
    assert_eq!(shake_offset(None, now), (0.0, 0.0));
    assert_eq!(shake_offset(Some(now), now + Duration::from_millis(1)), (0.0, 0.0));
    let (dx, dy) = shake_offset(Some(now + Duration::from_millis(200)), now);
    assert!(dx != 0.0 || dy != 0.0, "{}", l("medan det blinkar ska skärmen skaka", "while it flashes the screen should shake", "点滅中は画面が揺れるべきです"));
    assert!(dx.abs() <= 0.5 && dy.abs() <= 0.5, "{}", l("men bara lite", "but only a little", "ただし少しだけ"));
}

#[derive(Default)]
struct OnlyCells {
    cells: usize,
}

impl Renderer for OnlyCells {
    fn cell(&mut self, _x: u16, _y: u16, _cell: &SceneCell) {
        self.cells += 1;
    }
    fn text(&mut self, _x: u16, _y: u16, _text: &str, _fg: Rgb) {}
    // no particle(): the trait's default method is used
}

#[test]
fn particle_is_a_default_method() {
    let mut r = OnlyCells::default();
    r.particle(1.0, 1.0, Rgb(255, 0, 0), 1.0);
    assert_eq!(r.cells, 0);
    let _ = FOG;
}
