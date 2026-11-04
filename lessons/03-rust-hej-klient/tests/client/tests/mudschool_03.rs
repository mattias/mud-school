// This file belongs to the course tool and is overwritten by `mudschool check`.

mod mudschool_common;

use std::io::{BufRead, BufReader, Write};
use std::net::TcpListener;
use std::thread;

use mud_client::net::Connection;
use mudschool_common::l;

/// Starts a small fake server that greets and echoes `rounds` lines,
/// then closes the connection.
fn fake_server(rounds: usize) -> String {
    let listener = TcpListener::bind("127.0.0.1:0").unwrap();
    let addr = listener.local_addr().unwrap().to_string();
    thread::spawn(move || {
        let (stream, _) = listener.accept().unwrap();
        let mut writer = stream.try_clone().unwrap();
        let mut reader = BufReader::new(stream);
        writer
            .write_all("Welcome to MUD School!\r\n".as_bytes())
            .unwrap();
        for _ in 0..rounds {
            let mut line = String::new();
            if reader.read_line(&mut line).unwrap() == 0 {
                return;
            }
            assert!(
                line.ends_with("\r\n"),
                "{} {line:?}",
                l(
                    "klienten ska avsluta rader med \\r\\n, men skickade",
                    "the client should end lines with \\r\\n, but sent",
                    "クライアントは行を \\r\\n で終えるべきですが、送信内容は"
                )
            );
            writer.write_all(line.as_bytes()).unwrap();
        }
        // Dropping the streams closes the connection.
    });
    addr
}

#[test]
fn connects_and_reads_welcome() {
    let addr = fake_server(0);
    let mut conn = Connection::connect(&addr).expect(l(
        "connect ska lyckas",
        "connect should succeed",
        "connect は成功するべきです",
    ));
    let line = conn.read_line().expect(l(
        "read_line ska lyckas",
        "read_line should succeed",
        "read_line は成功するべきです",
    ));
    assert_eq!(line, Some("Welcome to MUD School!".to_string()));
}

#[test]
fn sends_and_receives_lines() {
    let addr = fake_server(3);
    let mut conn = Connection::connect(&addr).unwrap();
    conn.read_line().unwrap();
    for msg in ["hello", "åäö", "こんにちは"] {
        conn.send_line(msg).unwrap();
        assert_eq!(conn.read_line().unwrap().as_deref(), Some(msg));
    }
}

#[test]
fn read_line_returns_none_when_server_closes() {
    let addr = fake_server(0);
    let mut conn = Connection::connect(&addr).unwrap();
    conn.read_line().unwrap();
    assert_eq!(
        conn.read_line().unwrap(),
        None,
        "{}",
        l(
            "read_line ska ge Ok(None) när servern stängt",
            "read_line should return Ok(None) when the server has closed",
            "サーバーが閉じたら read_line は Ok(None) を返すべきです"
        )
    );
}

#[test]
fn connect_to_closed_port_is_an_error() {
    // Bind and release immediately -> the port is (almost certainly) closed.
    let addr = {
        let l = TcpListener::bind("127.0.0.1:0").unwrap();
        l.local_addr().unwrap().to_string()
    };
    assert!(
        Connection::connect(&addr).is_err(),
        "{}",
        l(
            "connect till en stängd port ska ge Err, inte panik",
            "connect to a closed port should return Err, not panic",
            "閉じたポートへの connect はパニックではなく Err を返すべきです"
        )
    );
}
