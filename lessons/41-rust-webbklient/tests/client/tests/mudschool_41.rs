// This file belongs to the course tool and is overwritten by `mudschool check`.

mod mudschool_common;

use std::path::{Path, PathBuf};
use std::process::Command;

use mud_web::{WebClient, ABI_VERSION, CELL_BYTES, WALK_MS};
use mudschool_common::l;

/// A 7×5 zone: Anna at (1,1), a rat at (5,3).
fn zone_line(room: &str) -> String {
    serde_json::json!({"type": "zone", "zone": {
        "room": room, "width": 7, "height": 5,
        "rows": ["#######", "#.....#", "#.....#", "#.....#", "#######"],
        "legend": {"#": "wall", ".": "floor"},
        "entities": [
            {"id": "p:Anna", "kind": "player", "name": "Anna", "glyph": "@", "x": 1, "y": 1},
            {"id": "n:rat", "kind": "creature", "name": "a rat", "glyph": "r", "x": 5, "y": 3}
        ],
        "self": "p:Anna"}})
    .to_string()
}

fn glyph_at(frame: &[u8], cols: usize, x: usize, y: usize) -> char {
    let o = (y * cols + x) * CELL_BYTES;
    char::from_u32(u32::from_le_bytes(frame[o..o + 4].try_into().unwrap())).unwrap_or('?')
}

fn find(frame: &[u8], cols: usize, glyph: char) -> Option<(usize, usize)> {
    (0..frame.len() / CELL_BYTES).find(|i| glyph_at(frame, cols, i % cols, i / cols) == glyph).map(|i| (i % cols, i / cols))
}

#[test]
fn an_empty_frame() {
    let mut c = WebClient::new();
    let f = c.frame(10, 4, 0);
    assert_eq!(f.len(), 10 * 4 * CELL_BYTES, "{}", l("en frame är cols × rows × CELL_BYTES byte", "a frame is cols × rows × CELL_BYTES bytes", "フレームは cols × rows × CELL_BYTES バイトです"));
    assert!(f.iter().all(|&b| b == 0), "{}", l("utan zon är alla byte 0", "with no zone every byte is 0", "ゾーンがなければすべて 0 です"));
}

#[test]
fn the_zone_in_a_frame() {
    let mut c = WebClient::new();
    c.on_line(&zone_line("r"));
    c.on_line("just some text"); // not ours: ignored
    let cols = 7;
    let f = c.frame(cols, 5, 0).to_vec();
    assert_eq!(glyph_at(&f, cols, 1, 1), '@', "{}", l("Anna ritas som @ på (1,1)", "Anna is drawn as @ at (1,1)", "アンナは (1,1) に @ として描かれます"));
    assert_eq!(glyph_at(&f, cols, 5, 3), 'r');
    let o = (cols + 1) * CELL_BYTES;
    assert!(f[o + 4..o + 7].iter().any(|&b| b > 0), "{}", l("@ har en färg (fg)", "the @ has a colour (fg)", "@ には色（fg）があります"));

    // A bigger canvas: the small zone is drawn in the middle.
    let f = c.frame(15, 9, 0).to_vec();
    assert_eq!(find(&f, 15, '@'), Some((5, 3)), "{}", l(
        "en zon som är mindre än canvasen ritas i mitten: (15-7)/2 = 4 kolumner och (9-5)/2 = 2 rader in",
        "a zone smaller than the canvas is drawn in the middle: (15-7)/2 = 4 columns and (9-5)/2 = 2 rows in",
        "キャンバスより小さいゾーンは中央に描きます: 左から (15-7)/2 = 4 列、上から (9-5)/2 = 2 行"));
}

#[test]
fn deltas_move_things() {
    let mut c = WebClient::new();
    c.on_line(&zone_line("r"));
    c.on_line(r#"{"type":"delta","delta":{"room":"r","moved":[{"id":"n:rat","kind":"creature","name":"a rat","glyph":"r","x":4,"y":2}]}}"#);
    let f = c.frame(7, 5, 0).to_vec();
    assert_eq!(find(&f, 7, 'r'), Some((4, 2)), "{}", l("råttan har flyttat till (4,2)", "the rat has moved to (4,2)", "ネズミは (4,2) に動きました"));
}

#[test]
fn the_light_depends_on_the_terrain() {
    let mut c = WebClient::new();
    c.on_line(&zone_line("r"));
    let fg = |c: &mut WebClient| {
        let f = c.frame(7, 5, 0);
        let o = (3 * 7 + 5) * CELL_BYTES; // the rat, far from Anna's torch
        f[o + 4] as u32 + f[o + 5] as u32 + f[o + 6] as u32
    };
    c.on_line(r#"{"type":"status","status":{"name":"Anna","room":"Square","terrain":"town","hp":10,"max_hp":10,"online":1}}"#);
    let town = fg(&mut c);
    c.on_line(r#"{"type":"status","status":{"name":"Anna","room":"Cave","terrain":"cave","hp":10,"max_hp":10,"online":1}}"#);
    let cave = fg(&mut c);
    assert!(cave < town, "{} ({cave} / {town})", l(
        "samma ruta är mörkare i en grotta än i en stad – använd zone_light från kärnan",
        "the same cell is darker in a cave than in a town – use zone_light from the core",
        "同じマスでも洞窟では町より暗くなります。コアの zone_light を使いましょう"));
}

#[test]
fn click_to_walk() {
    let mut c = WebClient::new();
    c.on_line(&zone_line("r"));
    c.frame(7, 5, 0);
    c.click(4, 1);
    let mut sent = vec![];
    let mut t = 1000.0;
    for _ in 0..10 {
        c.tick(t);
        c.tick(t + 1.0); // too soon: no second step
        sent.extend(c.take_outbox());
        t += WALK_MS;
    }
    assert_eq!(sent, ["east", "east", "east"], "{}", l(
        "ett klick på (4,1) går tre steg österut, ett steg per WALK_MS",
        "a click at (4,1) walks three steps east, one step per WALK_MS",
        "(4,1) をクリックすると、WALK_MS ごとに 1 歩ずつ東へ 3 歩進みます"));

    c.click(4, 3);
    c.on_line(&zone_line("another room"));
    c.tick(t + 1000.0);
    assert!(c.take_outbox().is_empty(), "{}", l(
        "en zon för ett annat rum avbryter promenaden",
        "a zone for another room stops the walk",
        "別の部屋のゾーンが来たら歩きは中止です"));
}

#[test]
fn the_abi_version() {
    assert_eq!(ABI_VERSION, 1);
    assert_eq!(CELL_BYTES, 12, "{}", l("mud.js läser 12 byte per ruta", "mud.js reads 12 bytes per cell", "mud.js はマスごとに 12 バイト読みます"));
}

// ---- the real thing: build for the browser and look inside ----------------

fn target_dir() -> PathBuf {
    let base = std::env::var_os("CARGO_TARGET_DIR")
        .map(PathBuf::from)
        .unwrap_or_else(|| Path::new(env!("CARGO_MANIFEST_DIR")).join("target"));
    base.join("wasm-check") // a directory of its own: the outer cargo holds a lock on the normal one
}

fn leb(b: &[u8], i: &mut usize) -> u32 {
    let (mut v, mut shift) = (0u32, 0);
    loop {
        let byte = b[*i];
        *i += 1;
        v |= ((byte & 0x7f) as u32) << shift;
        if byte & 0x80 == 0 {
            return v;
        }
        shift += 7;
    }
}

fn name(b: &[u8], i: &mut usize) -> String {
    let n = leb(b, i) as usize;
    let s = String::from_utf8_lossy(&b[*i..*i + n]).to_string();
    *i += n;
    s
}

/// The names a WebAssembly module imports ("module.field") and exports.
fn imports_and_exports(wasm: &[u8]) -> (Vec<String>, Vec<String>) {
    assert_eq!(&wasm[0..4], b"\0asm");
    let (mut imports, mut exports) = (vec![], vec![]);
    let mut i = 8;
    while i < wasm.len() {
        let id = wasm[i];
        i += 1;
        let size = leb(wasm, &mut i) as usize;
        let end = i + size;
        if id == 2 || id == 7 {
            let mut j = i;
            for _ in 0..leb(wasm, &mut j) {
                if id == 2 {
                    let module = name(wasm, &mut j);
                    let field = name(wasm, &mut j);
                    imports.push(format!("{module}.{field}"));
                    match wasm[j] {
                        0 => {
                            j += 1;
                            leb(wasm, &mut j);
                        }
                        kind => panic!("unexpected import kind {kind}"),
                    }
                } else {
                    exports.push(name(wasm, &mut j));
                    j += 1;
                    leb(wasm, &mut j);
                }
            }
        }
        i = end;
    }
    (imports, exports)
}

#[test]
fn it_builds_for_the_browser() {
    let cargo = std::env::var("CARGO").unwrap_or_else(|_| "cargo".into());
    let out = Command::new(cargo)
        .args(["build", "--release", "--quiet", "--target", "wasm32-unknown-unknown", "-p", "mud-web"])
        .env("CARGO_TARGET_DIR", target_dir())
        .current_dir(env!("CARGO_MANIFEST_DIR"))
        .output()
        .expect("cargo");
    let stderr = String::from_utf8_lossy(&out.stderr);
    assert!(
        out.status.success(),
        "{}\n{stderr}",
        if stderr.contains("wasm32-unknown-unknown") && stderr.contains("target may not be installed") {
            l(
                "WebAssembly-målet saknas: kör rustup target add wasm32-unknown-unknown",
                "the WebAssembly target is missing: run rustup target add wasm32-unknown-unknown",
                "WebAssembly のターゲットがありません: rustup target add wasm32-unknown-unknown を実行してください",
            )
        } else {
            l("mud-web byggs inte för wasm32-unknown-unknown", "mud-web doesn't build for wasm32-unknown-unknown", "mud-web が wasm32-unknown-unknown 向けにビルドできません")
        }
    );
    let wasm = std::fs::read(target_dir().join("wasm32-unknown-unknown/release/mud_web.wasm")).expect("mud_web.wasm");
    let (imports, exports) = imports_and_exports(&wasm);
    for want in ["memory", "abi_version", "alloc", "dealloc", "on_line", "frame", "click", "tick"] {
        assert!(exports.iter().any(|e| e == want), "{} {want:?}: {exports:?}", l("mud_web.wasm saknar exporten", "mud_web.wasm lacks the export", "mud_web.wasm にエクスポートがありません:"));
    }
    assert_eq!(imports, ["mud.send"], "{}", l(
        "mud_web.wasm ska bara importera mud.send – allt annat finns inte i sidan",
        "mud_web.wasm should only import mud.send – the page offers nothing else",
        "mud_web.wasm がインポートするのは mud.send だけです。ページはそれ以外を用意していません"));
}
