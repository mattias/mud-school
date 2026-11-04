// Den här filen hör till kursverktyget och skrivs över av `mudschool check`.

use std::collections::{HashMap, HashSet};

use mud_client::protocol::MapData;
use mud_client::render::{render_map, tile, to_ansi, Cell, Rgb, BLANK};

const TERRAINS: [&str; 8] = [
    "town", "inside", "road", "field", "forest", "water", "mountain", "cave",
];

fn test_map() -> MapData {
    let legend: HashMap<String, String> = [
        ("T", "forest"),
        ("~", "water"),
        ("@", "self"),
        ("-", "path_ew"),
        ("|", "path_ns"),
        ("#", "town"),
    ]
    .into_iter()
    .map(|(k, v)| (k.to_string(), v.to_string()))
    .collect();
    MapData {
        width: 5,
        height: 3,
        rows: vec!["T-~  ".into(), "|    ".into(), "#-@-X".into()],
        legend,
    }
}

#[test]
fn every_terrain_has_its_own_glyph() {
    let glyphs: HashSet<char> = TERRAINS.iter().map(|t| tile(t).glyph).collect();
    assert_eq!(
        glyphs.len(),
        TERRAINS.len(),
        "alla terränger ska ha olika tecken, men fick {glyphs:?}"
    );
    assert!(
        !glyphs.contains(&' '),
        "ingen terräng får vara ett mellanslag"
    );
}

#[test]
fn colors_make_sense() {
    let Rgb(r, g, b) = tile("forest").fg;
    assert!(g > r && g > b, "skog ska vara grön, är Rgb({r}, {g}, {b})");
    let water = tile("water");
    let Rgb(r, g, b) = water.bg.unwrap_or(water.fg);
    assert!(
        b > r && b > g,
        "vatten ska vara blått, är Rgb({r}, {g}, {b})"
    );
}

#[test]
fn self_is_an_at_sign_and_paths_are_different() {
    assert_eq!(tile("self").glyph, '@');
    assert_ne!(tile("path_ew").glyph, tile("path_ns").glyph);
    assert_ne!(tile("stairs_up").glyph, tile("stairs_down").glyph);
}

#[test]
fn unknown_meaning_does_not_panic() {
    let c = tile("något-helt-nytt");
    assert_ne!(c, tile("forest"));
    assert_ne!(c.glyph, ' ');
}

#[test]
fn render_map_builds_a_grid() {
    let map = test_map();
    let grid = render_map(&map);
    assert_eq!(grid.len(), 3, "en rad i rutnätet per rad i kartan");
    assert!(
        grid.iter().all(|row| row.len() == 5),
        "varje rad ska ha 5 celler"
    );
    assert_eq!(grid[0][0], tile("forest"));
    assert_eq!(grid[0][2], tile("water"));
    assert_eq!(grid[2][2], tile("self"));
    assert_eq!(grid[1][0], tile("path_ns"));
    assert_eq!(grid[0][4], BLANK);
    // 'X' finns inte i legenden: visa tecknet som det är.
    assert_eq!(grid[2][4].glyph, 'X');
}

/// Tar bort ANSI-koder: ESC [ ... m
fn strip_ansi(s: &str) -> String {
    let mut out = String::new();
    let mut chars = s.chars();
    while let Some(c) = chars.next() {
        if c == '\x1b' {
            for c in chars.by_ref() {
                if c == 'm' {
                    break;
                }
            }
        } else {
            out.push(c);
        }
    }
    out
}

#[test]
fn to_ansi_has_colors_and_the_right_text() {
    let grid = render_map(&test_map());
    let ansi = to_ansi(&grid);
    let Rgb(r, g, b) = tile("forest").fg;
    assert!(
        ansi.contains(&format!("\x1b[38;2;{r};{g};{b}m")),
        "skogen ska ha sin förgrundsfärg som truecolor-kod"
    );
    assert!(
        ansi.contains("\x1b[0m"),
        "färgerna ska återställas med \\x1b[0m"
    );

    let expected: Vec<String> = grid
        .iter()
        .map(|row| row.iter().map(|c: &Cell| c.glyph).collect())
        .collect();
    assert_eq!(strip_ansi(&ansi), expected.join("\n"));
}

#[test]
fn background_colors() {
    let grid = vec![vec![Cell {
        glyph: '≈',
        fg: Rgb(1, 2, 3),
        bg: Some(Rgb(4, 5, 6)),
    }]];
    let ansi = to_ansi(&grid);
    assert!(ansi.contains("\x1b[38;2;1;2;3m"));
    assert!(
        ansi.contains("\x1b[48;2;4;5;6m"),
        "bakgrundsfärg ska bli \\x1b[48;2;R;G;Bm"
    );
}
