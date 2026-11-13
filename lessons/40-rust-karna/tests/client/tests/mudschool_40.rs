// This file belongs to the course tool and is overwritten by `mudschool check`.

mod mudschool_common;

use std::fs;
use std::path::{Path, PathBuf};

use mud_core::light::{zone_light, FOG};
use mud_core::protocol::ServerMsg;
use mud_core::zone::ZoneView;
use mudschool_common::l;

fn core_dir() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("core")
}

/// A 5×3 zone: Anna at (1,1), water at (3,1), and (4,2) out of sight.
fn zone() -> ZoneView {
    let frame = serde_json::json!({"type": "zone", "zone": {
        "room": "r", "width": 5, "height": 3,
        "rows": ["#####", "#..~#", "#####"],
        "legend": {"#": "wall", ".": "floor", "~": "water"},
        "entities": [{"id": "p:Anna", "kind": "player", "name": "Anna", "glyph": "@", "x": 1, "y": 1}],
        "self": "p:Anna",
        "visible": ["11111", "11111", "11110"]}});
    match mud_client::protocol::parse_line(&frame.to_string()) {
        ServerMsg::Zone { zone } => ZoneView::from(zone),
        other => panic!("{}: {other:?}", l("väntade en zone-frame", "expected a zone frame", "zone フレームを期待しました")),
    }
}

#[test]
fn the_core_is_a_crate_of_its_own() {
    let manifest = fs::read_to_string(core_dir().join("Cargo.toml")).unwrap_or_else(|_| {
        panic!(
            "{}",
            l(
                "hittar inte client/core/Cargo.toml – skapa kärnan som ett eget paket (mud-core) i arbetsytan",
                "can't find client/core/Cargo.toml – create the core as a package of its own (mud-core) in the workspace",
                "client/core/Cargo.toml が見つかりません。コアをワークスペース内の独立したパッケージ（mud-core）として作りましょう"
            )
        )
    });
    let doc: toml::Table = manifest.parse().expect("core/Cargo.toml");
    let name = doc.get("package").and_then(|p| p.get("name")).and_then(|n| n.as_str());
    assert_eq!(name, Some("mud-core"), "{}", l("paketet ska heta mud-core", "the package should be called mud-core", "パッケージ名は mud-core にしましょう"));
    let deps: Vec<&String> = doc.get("dependencies").and_then(|d| d.as_table()).map(|t| t.keys().collect()).unwrap_or_default();
    for banned in ["ratatui", "crossterm", "macroquad", "clap", "toml"] {
        assert!(
            !deps.iter().any(|d| *d == banned),
            "{} {banned}",
            l(
                "kärnan får inte bero på terminalen, fönstret eller kommandoraden – ta bort",
                "the core must not depend on the terminal, the window or the command line – remove",
                "コアは端末・ウィンドウ・コマンドラインに依存してはいけません。削除しましょう:"
            )
        );
    }
}

#[test]
fn the_core_does_no_io() {
    let mut files = vec![];
    for entry in fs::read_dir(core_dir().join("src")).expect("core/src") {
        let path = entry.unwrap().path();
        if path.extension().is_some_and(|e| e == "rs") {
            files.push(path);
        }
    }
    assert!(files.len() >= 6, "{}", l(
        "kärnan ska ha protocol, zone, render, tileset, light och path",
        "the core should have protocol, zone, render, tileset, light and path",
        "コアには protocol・zone・render・tileset・light・path が必要です"));
    for path in files {
        let code = fs::read_to_string(&path).unwrap();
        for banned in ["std::net", "std::fs", "std::thread", "std::process", "Instant::now", "SystemTime::now", "stdin()"] {
            assert!(
                !code.contains(banned),
                "{}: {banned} ({})",
                path.display(),
                l(
                    "kärnan gör ingen I/O och läser ingen klocka – det kan inte webbläsarens WebAssembly heller",
                    "the core does no I/O and reads no clock – the browser's WebAssembly can't either",
                    "コアは I/O をせず時計も読みません。ブラウザの WebAssembly にもできないからです"
                )
            );
        }
    }
}

#[test]
fn the_client_uses_the_core() {
    // The same types, whichever path you take: the client re-exports the core.
    let msg: mud_core::protocol::ServerMsg = mud_client::protocol::parse_line("hello");
    let _: mud_client::protocol::ServerMsg = msg;
    let z: mud_client::zone::ZoneView = mud_core::zone::ZoneView::default();
    let _: Option<Vec<mud_core::path::Cell>> = mud_client::mouse::find_path(&z, (0, 0), (0, 0));
    let _: mud_client::render::Rgb = mud_core::render::Rgb(1, 2, 3);
    let _: mud_core::light::Light = mud_client::light::Light { x: 0.0, y: 0.0, radius: 1.0, color: mud_core::light::TORCH };
}

#[test]
fn zone_light_in_the_core() {
    let z = zone();
    assert_eq!(zone_light(&z, "town", 4, 2, 0), FOG, "{}", l(
        "en ruta man inte ser ligger i dimma (FOG)",
        "a cell you can't see is in the fog (FOG)",
        "見えないマスは霧（FOG）の中です"));
    let near = zone_light(&z, "cave", 2, 1, 0);
    let ambient = mud_core::light::ambient("cave");
    assert!(near.0 > ambient.0, "{}: {near:?} / {ambient:?}", l(
        "facklan lyser upp rutan bredvid Anna mer än grottans ljus",
        "the torch lights the cell next to Anna more than the cave's light",
        "たいまつはアンナの隣のマスを洞窟の光より明るくします"));
    let a = zone_light(&z, "town", 3, 1, 0);
    let b = zone_light(&z, "town", 3, 1, 600);
    assert_ne!(a, b, "{}", l(
        "vatten skimrar: ljuset ändras med tiden (t_ms)",
        "water shimmers: the light changes with time (t_ms)",
        "水はきらめきます。光は時間（t_ms）で変わります"));
}

#[test]
fn the_status_line_is_the_clients() {
    let msg = mud_client::protocol::parse_line(
        r#"{"type":"status","status":{"name":"Anna","room":"Square","terrain":"town","hp":7,"max_hp":10,"online":2}}"#,
    );
    let ServerMsg::Status { status } = msg else {
        panic!("status");
    };
    let line = mud_client::scene::status_line(&status);
    for want in ["Anna", "Square", "7/10", mud_client::messages::ONLINE] {
        assert!(line.contains(want), "status_line: {line:?} {} {want:?}", l("saknar", "lacks", "に含まれない:"));
    }
}
