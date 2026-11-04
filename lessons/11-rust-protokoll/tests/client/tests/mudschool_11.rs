// Den här filen hör till kursverktyget och skrivs över av `mudschool check`.

use mud_client::protocol::{parse_line, MapData, ServerMsg, Status, HELLO};

#[test]
fn hello_constant() {
    assert!(HELLO.starts_with("@hello"), "HELLO ska börja med @hello");
}

#[test]
fn text_frame() {
    let msg = parse_line(r#"{"type":"text","text":"Torget\nEtt torg."}"#);
    assert_eq!(
        msg,
        ServerMsg::Text {
            text: "Torget\nEtt torg.".to_string()
        }
    );
}

#[test]
fn text_frame_without_text_is_empty_text() {
    // Servern använder omitempty, så en tom text skickas utan "text"-fält.
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
        panic!("väntade ServerMsg::Map");
    };
    assert_eq!((map.width, map.height), (3, 1));
    assert_eq!(map.rows, vec!["@-T".to_string()]);
    assert_eq!(map.meaning('T'), Some("forest"));
    assert_eq!(map.meaning('@'), Some("self"));
    assert_eq!(map.meaning('?'), None);
}

#[test]
fn status_frame() {
    let line = r#"{"type":"status","status":{"name":"Anna","room":"Torget","terrain":"town","hp":7,"max_hp":10,"online":3}}"#;
    let ServerMsg::Status { status } = parse_line(line) else {
        panic!("väntade ServerMsg::Status");
    };
    assert_eq!(
        status,
        Status {
            name: "Anna".into(),
            room: "Torget".into(),
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
        room: "Torget".into(),
        terrain: "town".into(),
        hp: 7,
        max_hp: 10,
        online: 3,
    };
    let shown = status.to_string(); // to_string kommer gratis med Display
    for part in ["Anna", "Torget", "7/10", "3"] {
        assert!(shown.contains(part), "{shown:?} ska innehålla {part:?}");
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
    for line in ["Välkommen till MUD School!", "{trasig json", ""] {
        assert_eq!(
            parse_line(line),
            ServerMsg::Text {
                text: line.to_string()
            },
            "raden {line:?} ska bli vanlig text"
        );
    }
}

#[test]
fn map_data_has_default() {
    // Default kommer att behövas i nästa lektion.
    let empty = MapData::default();
    assert!(empty.rows.is_empty());
}
