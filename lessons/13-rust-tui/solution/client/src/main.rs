// Entry point of your MUD client. Run it with: ./mudschool run client [address]

use std::io;
use std::sync::mpsc;
use std::thread;

use ratatui::crossterm::event::{self, Event as TermEvent, KeyEvent};

use mud_client::app::{draw, App};
use mud_client::messages;
use mud_client::net::Connection;
use mud_client::protocol::{parse_line, ServerMsg, HELLO};

/// Everything that can happen – from the keyboard or the server – in one type.
enum Event {
    Key(KeyEvent),
    Server(ServerMsg),
    Disconnected,
}

fn main() -> io::Result<()> {
    let addr = std::env::args()
        .nth(1)
        .unwrap_or_else(|| "127.0.0.1:4000".to_string());
    let conn = Connection::connect(&addr)?;
    let (reader, mut outgoing) = conn.split();
    outgoing.send_line(HELLO)?;

    let (tx, rx) = mpsc::channel();

    // Thread 1: read from the server.
    let server_tx = tx.clone();
    thread::spawn(move || {
        use std::io::BufRead;
        for line in reader.lines() {
            let Ok(line) = line else { break };
            if server_tx.send(Event::Server(parse_line(&line))).is_err() {
                return;
            }
        }
        let _ = server_tx.send(Event::Disconnected);
    });

    // Thread 2: read the keyboard.
    thread::spawn(move || loop {
        match event::read() {
            Ok(TermEvent::Key(key)) => {
                if tx.send(Event::Key(key)).is_err() {
                    return;
                }
            }
            Ok(_) => {} // e.g. a resized window – we redraw anyway
            Err(_) => return,
        }
    });

    let mut terminal = ratatui::init();
    let mut app = App::new();
    app.push_log(&format!("{} {addr}", messages::CONNECTED_TO));

    // The main loop: draw, wait for an event, update.
    let result = (|| -> io::Result<()> {
        while !app.quit {
            terminal.draw(|f| draw(f, &app))?;
            let Ok(ev) = rx.recv() else { break };
            for ev in std::iter::once(ev).chain(rx.try_iter()) {
                match ev {
                    Event::Key(key) => {
                        if let Some(line) = app.on_key(key) {
                            outgoing.send_line(&line)?;
                        }
                    }
                    Event::Server(msg) => app.on_server(msg),
                    Event::Disconnected => app.on_disconnect(),
                }
            }
        }
        Ok(())
    })();

    ratatui::restore();
    result
}
