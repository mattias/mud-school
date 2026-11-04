// This file belongs to the course tool and is overwritten by `mudschool check`.

mod mudschool_common;

use mud_client::protocol::{parse_line, MapData, ServerMsg, Status, HELLO};
use mudschool_common::l;

#[test]
fn hello_constant() {
    assert!(
        HELLO.starts_with("@hello"),
        "{}",
        l(
            "HELLO ska börja med @hello",
            "HELLO should start with @hello",
            "HELLO は @hello で始まるべきです"
        )
    );
}

#[test]
fn text_frame() {
    let msg = parse_line(r#"{"type":"text","text":"Square\nA square."}"#);
    assert_eq!(
        msg,
        ServerMsg::Text {
            text: "Square\nA square.".to_string()
        }
    );
}

#[test]
fn text_frame_without_text_is_empty_text() {
    // The server uses omitempty, so an empty text is sent without a "text" field.
    assert_eq!(
        parse_line(r#"{"type":"text"}"#),
        ServerMsg::Text {
            text: String::new()
        }
    );
}

#[test]
fn hello_frame() {
    assert!(matches!(
        parse_line(r#"{"type":"hello","text":"mudschool 1"}"#),
        ServerMsg::Hello { .. }
    ));
}

#[test]
fn map_frame() {
    let line = r#"{"type":"map","map":{"width":3,"height":1,"rows":["@-T"],"legend":{"@":"self","-":"path_ew","T":"forest"}}}"#;
    let ServerMsg::Map { map } = parse_line(line) else {
        panic!(
            "{}",
            l(
                "väntade ServerMsg::Map",
                "expected ServerMsg::Map",
                "ServerMsg::Map を期待しました"
            )
        );
    };
    assert_eq!((map.width, map.height), (3, 1));
    assert_eq!(map.rows, vec!["@-T".to_string()]);
    assert_eq!(map.meaning('T'), Some("forest"));
    assert_eq!(map.meaning('@'), Some("self"));
    assert_eq!(map.meaning('?'), None);
}

#[test]
fn status_frame() {
    let line = r#"{"type":"status","status":{"name":"Anna","room":"Square","terrain":"town","hp":7,"max_hp":10,"online":3}}"#;
    let ServerMsg::Status { status } = parse_line(line) else {
        panic!(
            "{}",
            l(
                "väntade ServerMsg::Status",
                "expected ServerMsg::Status",
                "ServerMsg::Status を期待しました"
            )
        );
    };
    assert_eq!(
        status,
        Status {
            name: "Anna".into(),
            room: "Square".into(),
            terrain: "town".into(),
            hp: 7,
            max_hp: 10,
            online: 3,
        }
    );
}

#[test]
fn status_display() {
    let status = Status {
        name: "Anna".into(),
        room: "Square".into(),
        terrain: "town".into(),
        hp: 7,
        max_hp: 10,
        online: 3,
    };
    let shown = status.to_string(); // to_string comes for free with Display
    for part in ["Anna", "Square", "7/10", "3"] {
        assert!(
            shown.contains(part),
            "{shown:?} {} {part:?}",
            l("ska innehålla", "should contain", "に含まれるべき:")
        );
    }
}

#[test]
fn unknown_frame_type_is_not_an_error() {
    assert_eq!(
        parse_line(r#"{"type":"sound","sound":"thunder.ogg"}"#),
        ServerMsg::Unknown
    );
}

#[test]
fn plain_text_and_broken_json_become_text() {
    for line in ["Welcome to MUD School!", "{broken json", ""] {
        assert_eq!(
            parse_line(line),
            ServerMsg::Text {
                text: line.to_string()
            },
            "{line:?}: {}",
            l(
                "ska bli vanlig text",
                "should become plain text",
                "は普通のテキストになるべきです"
            )
        );
    }
}

#[test]
fn map_data_has_default() {
    // Default will be needed in the next lesson.
    let empty = MapData::default();
    assert!(empty.rows.is_empty());
}
