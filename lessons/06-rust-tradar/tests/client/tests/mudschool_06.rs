// This file belongs to the course tool and is overwritten by `mudschool check`.

mod mudschool_common;

use std::io::{BufRead, BufReader, Cursor, Write};
use std::net::TcpListener;
use std::sync::mpsc::{self, RecvTimeoutError};
use std::thread;
use std::time::Duration;

use mud_client::net::Connection;
use mud_client::reader::spawn_reader;
use mudschool_common::l;

const TIMEOUT: Duration = Duration::from_secs(2);

fn expect_line() -> &'static str {
    l(
        "väntade en rad på kanalen",
        "expected a line on the channel",
        "チャネルに行が届くはずです",
    )
}

#[test]
fn reads_all_lines_without_newlines() {
    let rx = spawn_reader(Cursor::new("first\r\nåäö\n日本語".as_bytes().to_vec()));
    let mut got = Vec::new();
    for _ in 0..3 {
        got.push(rx.recv_timeout(TIMEOUT).expect(expect_line()));
    }
    assert_eq!(got, ["first", "åäö", "日本語"]);
}

#[test]
fn channel_closes_at_end_of_input() {
    let rx = spawn_reader(Cursor::new(b"one line\n".to_vec()));
    assert_eq!(rx.recv_timeout(TIMEOUT).expect(expect_line()), "one line");
    assert_eq!(
        rx.recv_timeout(TIMEOUT),
        Err(RecvTimeoutError::Disconnected),
        "{}",
        l(
            "kanalen ska stängas när det inte finns mer att läsa",
            "the channel should close when there is nothing more to read",
            "読むものがなくなったらチャネルは閉じるべきです"
        )
    );
}

#[test]
fn works_with_any_bufread() {
    // &'static [u8] implements BufRead too – generics in practice.
    let rx = spawn_reader(&b"hi\n"[..]);
    assert_eq!(rx.recv_timeout(TIMEOUT).expect(expect_line()), "hi");
}

#[test]
fn server_can_talk_first() {
    // A server that talks without waiting for us – just like when another
    // player says something.
    let listener = TcpListener::bind("127.0.0.1:0").unwrap();
    let addr = listener.local_addr().unwrap().to_string();
    let (got_tx, got_rx) = mpsc::channel();
    thread::spawn(move || {
        let (mut stream, _) = listener.accept().unwrap();
        stream
            .write_all(b"Welcome!\r\nAnna has connected.\r\nAnna says: hi\r\n")
            .unwrap();
        let mut line = String::new();
        BufReader::new(stream).read_line(&mut line).unwrap();
        got_tx.send(line).unwrap();
    });

    let conn = Connection::connect(&addr).unwrap();
    let (reader, mut outgoing) = conn.split();
    let rx = spawn_reader(reader);
    for want in ["Welcome!", "Anna has connected.", "Anna says: hi"] {
        assert_eq!(rx.recv_timeout(TIMEOUT).expect(expect_line()), want);
    }
    outgoing.send_line("hi yourself").unwrap();
    assert_eq!(got_rx.recv_timeout(TIMEOUT).unwrap(), "hi yourself\r\n");
}
