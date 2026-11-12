// This file belongs to the course tool and is overwritten by `mudschool check`.

mod mudschool_common;

use std::collections::HashSet;
use std::io::{BufRead, BufReader, Write};
use std::net::{TcpListener, TcpStream};
use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::Arc;
use std::thread;
use std::time::Duration;

use mud_loadtest::{bot_name, percentile, run_blocking, Plan, Report};
use mudschool_common::l;

/// A tiny fake MUD with plain threads: it asks for a name, answers the
/// handshake with JSON, sends a zone after the name, and answers "echo X"
/// with X. It counts the commands it gets.
fn fake_server() -> (String, Arc<AtomicUsize>) {
    let listener = TcpListener::bind("127.0.0.1:0").unwrap();
    let addr = listener.local_addr().unwrap().to_string();
    let commands = Arc::new(AtomicUsize::new(0));
    let counter = commands.clone();
    thread::spawn(move || {
        for conn in listener.incoming().flatten() {
            let counter = counter.clone();
            thread::spawn(move || serve(conn, counter));
        }
    });
    (addr, commands)
}

fn serve(conn: TcpStream, commands: Arc<AtomicUsize>) {
    let mut w = conn.try_clone().unwrap();
    let _ = w.write_all(b"What is your name?\r\n");
    let mut named = false;
    for line in BufReader::new(conn).lines() {
        let Ok(line) = line else { return };
        let reply = if line.starts_with("@hello") {
            r#"{"type":"hello","text":"fake 1"}"#.to_string()
        } else if !named {
            named = true;
            serde_json::json!({"type": "zone", "zone": {
                "room": "r", "width": 5, "height": 3, "rows": ["#####", "#...#", "#####"],
                "legend": {"#": "wall", ".": "floor"},
                "entities": [{"id": format!("p:{line}"), "kind": "player", "name": line, "glyph": "@", "x": 2, "y": 1}],
                "self": format!("p:{line}")}})
            .to_string()
        } else {
            commands.fetch_add(1, Ordering::SeqCst);
            let text = line.strip_prefix("echo ").unwrap_or("ok");
            serde_json::json!({"type": "text", "kind": "info", "text": text}).to_string()
        };
        if w.write_all(format!("{reply}\r\n").as_bytes()).is_err() {
            return;
        }
    }
}

#[test]
fn percentiles() {
    let v: Vec<f64> = (1..=100).map(|i| i as f64).collect();
    assert_eq!(percentile(&v, 50.0), 50.0);
    assert_eq!(percentile(&v, 95.0), 95.0);
    assert_eq!(percentile(&v, 100.0), 100.0);
    assert_eq!(percentile(&[7.0], 99.0), 7.0);
    assert_eq!(percentile(&[], 50.0), 0.0, "{}", l("inga värden: 0", "no values: 0", "値がなければ 0"));
}

#[test]
fn bot_names_are_valid_and_unique() {
    let mut seen = HashSet::new();
    for id in 0..2000 {
        let name = bot_name(id);
        assert!(
            name.chars().all(|c| c.is_ascii_alphabetic()) && (2..=16).contains(&name.len()),
            "{}: {name:?}",
            l("servern tar bara namn med 2–16 bokstäver", "the server only accepts names of 2–16 letters", "サーバーが受け付けるのは 2〜16 文字の英字の名前だけです")
        );
        assert!(seen.insert(name.clone()), "{}: {name:?}", l("två botar fick samma namn", "two bots got the same name", "2 体のボットが同じ名前です"));
    }
}

#[test]
fn many_bots_at_once() {
    let (addr, commands) = fake_server();
    let report = run_blocking(Plan {
        addr,
        bots: 40,
        duration: Duration::from_millis(1500),
        think: Duration::from_millis(50),
        ramp: Duration::from_millis(200),
    });
    assert_eq!((report.connected, report.failed, report.logged_in), (40, 0, 40), "{}\n{report}", l(
        "alla 40 botar ska ansluta och logga in (en zone-frame = inloggad)",
        "all 40 bots should connect and log in (a zone frame = logged in)",
        "40 体すべてのボットが接続してログインするべきです（zone フレーム = ログイン済み）"));
    assert!(report.commands >= 200, "{}\n{report}", l(
        "40 botar som skickar ett kommando var ~50–75 ms i över en sekund ska skicka minst 200",
        "40 bots sending a command every ~50–75 ms for over a second should send at least 200",
        "約 50〜75 ms ごとに 1 秒以上コマンドを送る 40 体なら、少なくとも 200 個送るはずです"));
    assert!(report.replies > 0 && report.latencies_ms.len() == report.replies, "{}\n{report}", l(
        "svaren på echo ska mätas",
        "the answers to echo should be measured",
        "echo への応答時間を測るべきです"));
    assert!(report.latencies_ms.windows(2).all(|w| w[0] <= w[1]), "{}", l("svarstiderna ska vara sorterade", "the answer times should be sorted", "応答時間はソート済みであるべきです"));
    assert!(commands.load(Ordering::SeqCst) >= report.commands * 9 / 10, "{}", l(
        "servern ska ha fått kommandona botarna räknade",
        "the server should have received the commands the bots counted",
        "ボットが数えたコマンドはサーバーに届いているべきです"));
    let text = report.to_string();
    assert!(text.contains("40") && text.contains("p95"), "{}: {text}", l("rapporten ska visa siffrorna", "the report should show the numbers", "レポートには数字が必要です"));
}

#[test]
fn a_server_that_isnt_there() {
    let port = TcpListener::bind("127.0.0.1:0").unwrap().local_addr().unwrap().port(); // free again
    let report: Report = run_blocking(Plan {
        addr: format!("127.0.0.1:{port}"),
        bots: 5,
        duration: Duration::from_millis(500),
        think: Duration::from_millis(50),
        ramp: Duration::ZERO,
    });
    assert_eq!((report.connected, report.failed), (0, 5), "{}\n{report}", l(
        "utan server misslyckas alla anslutningar – och testet hänger inte",
        "without a server every connection fails – and the test doesn't hang",
        "サーバーがなければすべての接続が失敗します。そしてテストは止まりません"));
    assert!(!report.errors.is_empty(), "{}", l("felet ska synas i rapporten", "the error should show in the report", "エラーがレポートに出るべきです"));
}
