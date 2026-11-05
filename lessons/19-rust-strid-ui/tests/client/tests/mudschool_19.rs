// This file belongs to the course tool and is overwritten by `mudschool check`.

mod mudschool_common;

use std::time::{Duration, Instant};

use ratatui::backend::TestBackend;
use ratatui::buffer::Buffer;
use ratatui::Terminal;

use mud_client::app::{draw_at, App, FLASH, FLASH_BG};
use mud_client::protocol::{parse_line, ServerMsg, Status, Target};
use mudschool_common::l;

fn status(hp: i32, target: Option<Target>) -> ServerMsg {
    ServerMsg::Status {
        status: Status {
            name: "Anna".into(),
            room: "Lair".into(),
            terrain: "cave".into(),
            hp,
            max_hp: 20,
            online: 1,
            target,
        },
    }
}

fn golem(hp: i32) -> Option<Target> {
    Some(Target {
        name: "a golem".into(),
        hp,
        max_hp: 40,
    })
}

fn render(app: &App, now: Instant) -> Buffer {
    let mut terminal = Terminal::new(TestBackend::new(80, 20)).unwrap();
    terminal.draw(|f| draw_at(f, app, now)).unwrap();
    terminal.backend().buffer().clone()
}

fn screen(buf: &Buffer) -> String {
    let a = buf.area;
    (0..a.height)
        .map(|y| {
            (0..a.width)
                .map(|x| buf[(x, y)].symbol())
                .collect::<String>()
        })
        .collect::<Vec<_>>()
        .join("\n")
}

#[test]
fn parses_target() {
    let with = parse_line(
        r#"{"type":"status","status":{"name":"Anna","room":"Lair","hp":17,"max_hp":20,"online":1,"target":{"name":"a golem","hp":31,"max_hp":40}}}"#,
    );
    let ServerMsg::Status { status } = with else {
        panic!("status")
    };
    assert_eq!(status.target, golem(31));

    let without = parse_line(
        r#"{"type":"status","status":{"name":"Anna","room":"Lair","hp":20,"max_hp":20}}"#,
    );
    let ServerMsg::Status { status } = without else {
        panic!("status")
    };
    assert_eq!(
        status.target,
        None,
        "{}",
        l(
            "utan \"target\" ska target vara None",
            "without \"target\" the target should be None",
            "\"target\" がなければ target は None であるべきです"
        )
    );
}

#[test]
fn flashes_when_hp_drops() {
    let t0 = Instant::now();
    let mut app = App::new();
    app.on_server_at(status(20, None), t0);
    assert!(
        !app.is_flashing(t0),
        "{}",
        l(
            "första statusen ska inte blinka",
            "the first status should not flash",
            "最初の status では点滅しないべきです"
        )
    );

    app.on_server_at(status(17, golem(40)), t0);
    assert!(
        app.is_flashing(t0 + Duration::from_millis(10)),
        "{}",
        l(
            "när hp minskar ska statusraden blinka",
            "when hp goes down the status bar should flash",
            "hp が減ったらステータスバーは点滅するべきです"
        )
    );
    assert!(
        !app.is_flashing(t0 + FLASH + Duration::from_millis(1)),
        "{}",
        l(
            "blinkningen ska sluta efter FLASH",
            "the flash should end after FLASH",
            "点滅は FLASH の後に終わるべきです"
        )
    );

    let t1 = t0 + Duration::from_secs(5);
    app.on_server_at(status(20, None), t1);
    assert!(
        !app.is_flashing(t1),
        "{}",
        l(
            "när hp ökar ska den inte blinka",
            "when hp goes up it should not flash",
            "hp が増えたときは点滅しないべきです"
        )
    );
}

#[test]
fn status_bar_turns_red_while_flashing() {
    let t0 = Instant::now();
    let mut app = App::new();
    app.on_server_at(status(20, None), t0);
    app.on_server_at(status(15, golem(40)), t0);
    let buf = render(&app, t0);
    assert_eq!(
        buf[(0, 0)].bg,
        FLASH_BG,
        "{}",
        l(
            "statusraden ska ha bakgrunden FLASH_BG medan den blinkar",
            "the status bar should have the background FLASH_BG while it flashes",
            "点滅中のステータスバーの背景は FLASH_BG であるべきです"
        )
    );
    let later = render(&app, t0 + FLASH * 2);
    assert_ne!(later[(0, 0)].bg, FLASH_BG);
}

#[test]
fn target_panel() {
    let t0 = Instant::now();
    let mut app = App::new();
    app.on_server_at(status(20, golem(10)), t0);
    let s = screen(&render(&app, t0));
    for want in ["a golem", "10/40"] {
        assert!(
            s.contains(want),
            "{} {want:?}:\n{s}",
            l(
                "målpanelen ska visa",
                "the target panel should show",
                "ターゲットパネルに表示されるべき:"
            )
        );
    }
    app.on_server_at(status(20, None), t0);
    let s = screen(&render(&app, t0));
    assert!(
        !s.contains("a golem"),
        "{}\n{s}",
        l(
            "utan target ska målpanelen försvinna",
            "without a target the panel should disappear",
            "target がなければパネルは消えるべきです"
        )
    );
}

#[test]
fn dead_target_shows_zero() {
    let t0 = Instant::now();
    let mut app = App::new();
    app.on_server_at(status(20, golem(-3)), t0);
    let s = screen(&render(&app, t0));
    assert!(s.contains("0/40"), "{s}");
}
