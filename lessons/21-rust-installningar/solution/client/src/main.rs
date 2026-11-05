// Entry point of your MUD client. Run it with: ./mudschool run client [address]
// See all the options with: ./mudschool run client --help

use std::io;
use std::process::ExitCode;
use std::sync::mpsc;
use std::thread;
use std::time::Duration;

use ratatui::crossterm::event::{self, Event as TermEvent, KeyEvent};

use clap::Parser;

use mud_client::app::{draw, App};
use mud_client::config::{Cli, Config};
use mud_client::messages;
use mud_client::net::Connection;
use mud_client::protocol::{parse_line, Kind, ServerMsg};

/// Everything that can happen – from the keyboard or the server – in one type.
enum Event {
    Key(KeyEvent),
    Server(ServerMsg),
    Disconnected,
}

fn main() -> ExitCode {
    // clap prints --help, --version and wrong arguments itself, and exits.
    let cli = Cli::parse();
    let config = match Config::load(&cli) {
        Ok(config) => config,
        Err(err) => {
            eprintln!("{}: {err}", messages::CONFIG_ERROR);
            return ExitCode::FAILURE;
        }
    };
    match run(&config) {
        Ok(()) => ExitCode::SUCCESS,
        Err(err) => {
            eprintln!("{err}");
            ExitCode::FAILURE
        }
    }
}

fn run(config: &Config) -> io::Result<()> {
    let addr = &config.server;
    let conn = Connection::connect(addr)?;
    let (reader, mut outgoing) = conn.split();
    for line in config.startup_lines() {
        outgoing.send_line(&line)?;
    }

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
    app.push_log(Kind::Info, &format!("{} {addr}", messages::CONNECTED_TO));

    // The main loop: draw, wait for an event, update.
    let result = (|| -> io::Result<()> {
        while !app.quit {
            terminal.draw(|f| draw(f, &app))?;
            // Wake up now and then even without events, so that time-based
            // effects (the flash) end on time.
            let ev = match rx.recv_timeout(Duration::from_millis(50)) {
                Ok(ev) => ev,
                Err(mpsc::RecvTimeoutError::Timeout) => continue,
                Err(mpsc::RecvTimeoutError::Disconnected) => break,
            };
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
