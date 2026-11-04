// Den här filen hör till kursverktyget och skrivs över av `mudschool check`.

use std::io::{BufRead, BufReader, Cursor, Write};
use std::net::TcpListener;
use std::sync::mpsc::{self, RecvTimeoutError};
use std::thread;
use std::time::Duration;

use mud_client::net::Connection;
use mud_client::reader::spawn_reader;

const TIMEOUT: Duration = Duration::from_secs(2);

#[test]
fn reads_all_lines_without_newlines() {
    let rx = spawn_reader(Cursor::new("första\r\nandra\ntredje".as_bytes().to_vec()));
    let mut got = Vec::new();
    for _ in 0..3 {
        got.push(rx.recv_timeout(TIMEOUT).expect("väntade en rad på kanalen"));
    }
    assert_eq!(got, ["första", "andra", "tredje"]);
}

#[test]
fn channel_closes_at_end_of_input() {
    let rx = spawn_reader(Cursor::new(b"en rad\n".to_vec()));
    assert_eq!(rx.recv_timeout(TIMEOUT).unwrap(), "en rad");
    assert_eq!(
        rx.recv_timeout(TIMEOUT),
        Err(RecvTimeoutError::Disconnected),
        "kanalen ska stängas när det inte finns mer att läsa"
    );
}

#[test]
fn works_with_any_bufread() {
    // &'static [u8] implementerar också BufRead – generics i praktiken.
    let rx = spawn_reader(&b"hej\n"[..]);
    assert_eq!(rx.recv_timeout(TIMEOUT).unwrap(), "hej");
}

#[test]
fn server_can_talk_first() {
    // En server som pratar utan att vänta på oss – precis som när någon
    // annan spelare säger något.
    let listener = TcpListener::bind("127.0.0.1:0").unwrap();
    let addr = listener.local_addr().unwrap().to_string();
    let (got_tx, got_rx) = mpsc::channel();
    thread::spawn(move || {
        let (mut stream, _) = listener.accept().unwrap();
        stream
            .write_all(b"Valkommen!\r\nAnna har anslutit.\r\nAnna sager: hej\r\n")
            .unwrap();
        let mut line = String::new();
        BufReader::new(stream).read_line(&mut line).unwrap();
        got_tx.send(line).unwrap();
    });

    let conn = Connection::connect(&addr).unwrap();
    let (reader, mut outgoing) = conn.split();
    let rx = spawn_reader(reader);
    for want in ["Valkommen!", "Anna har anslutit.", "Anna sager: hej"] {
        assert_eq!(rx.recv_timeout(TIMEOUT).unwrap(), want);
    }
    outgoing.send_line("hej själv").unwrap();
    assert_eq!(got_rx.recv_timeout(TIMEOUT).unwrap(), "hej själv\r\n");
}
