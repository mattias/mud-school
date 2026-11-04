// Den här filen hör till kursverktyget och skrivs över av `mudschool check`.

use std::io::{BufRead, BufReader, Write};
use std::net::TcpListener;
use std::thread;

use mud_client::net::Connection;

/// Startar en liten låtsasserver som hälsar och ekar `rounds` rader,
/// sedan stänger den anslutningen.
fn fake_server(rounds: usize) -> String {
    let listener = TcpListener::bind("127.0.0.1:0").unwrap();
    let addr = listener.local_addr().unwrap().to_string();
    thread::spawn(move || {
        let (stream, _) = listener.accept().unwrap();
        let mut writer = stream.try_clone().unwrap();
        let mut reader = BufReader::new(stream);
        writer
            .write_all("Välkommen till MUD School!\r\n".as_bytes())
            .unwrap();
        for _ in 0..rounds {
            let mut line = String::new();
            if reader.read_line(&mut line).unwrap() == 0 {
                return;
            }
            assert!(
                line.ends_with("\r\n"),
                "klienten ska avsluta rader med \\r\\n, men skickade {line:?}"
            );
            writer.write_all(line.as_bytes()).unwrap();
        }
        // Släpp strömmarna -> anslutningen stängs.
    });
    addr
}

#[test]
fn connects_and_reads_welcome() {
    let addr = fake_server(0);
    let mut conn = Connection::connect(&addr).expect("connect ska lyckas");
    let line = conn.read_line().expect("read_line ska lyckas");
    assert_eq!(line, Some("Välkommen till MUD School!".to_string()));
}

#[test]
fn sends_and_receives_lines() {
    let addr = fake_server(2);
    let mut conn = Connection::connect(&addr).unwrap();
    conn.read_line().unwrap();
    conn.send_line("hej").unwrap();
    assert_eq!(conn.read_line().unwrap().as_deref(), Some("hej"));
    conn.send_line("åäö").unwrap();
    assert_eq!(conn.read_line().unwrap().as_deref(), Some("åäö"));
}

#[test]
fn read_line_returns_none_when_server_closes() {
    let addr = fake_server(0);
    let mut conn = Connection::connect(&addr).unwrap();
    conn.read_line().unwrap();
    assert_eq!(conn.read_line().unwrap(), None);
}

#[test]
fn connect_to_closed_port_is_an_error() {
    // Bind och släpp direkt -> porten är (nästan säkert) stängd.
    let addr = {
        let l = TcpListener::bind("127.0.0.1:0").unwrap();
        l.local_addr().unwrap().to_string()
    };
    assert!(
        Connection::connect(&addr).is_err(),
        "connect till en stängd port ska ge Err, inte panik"
    );
}
