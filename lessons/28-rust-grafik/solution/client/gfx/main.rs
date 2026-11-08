// The graphical MUD client: the same App, protocol and network code as the
// terminal client, drawn with tiles in a window of its own.
// Run it with: ./mudschool run gfx [address] [--name NAME]

use std::io::BufRead;
use std::sync::mpsc;
use std::thread;

use clap::Parser;
use macroquad::prelude::*;
use ratatui::crossterm::event::{KeyCode as Key, KeyEvent, KeyModifiers};

use mud_client::app::App;
use mud_client::config::{Cli, Config};
use mud_client::gfx::GfxRenderer;
use mud_client::messages;
use mud_client::net::Connection;
use mud_client::protocol::{parse_line, Kind};
use mud_client::scene::{build_scene, draw_scene};

fn window_conf() -> Conf {
    Conf {
        window_title: "MUD School".to_string(),
        window_width: 1024,
        window_height: 720,
        ..Default::default()
    }
}

/// The keys the App understands, from macroquad's keyboard.
fn keys() -> Vec<KeyEvent> {
    let mut events: Vec<KeyEvent> = Vec::new();
    while let Some(c) = get_char_pressed() {
        if !c.is_control() {
            events.push(KeyEvent::new(Key::Char(c), KeyModifiers::NONE));
        }
    }
    let map = [
        (KeyCode::Enter, Key::Enter),
        (KeyCode::Backspace, Key::Backspace),
        (KeyCode::Escape, Key::Esc),
        (KeyCode::Up, Key::Up),
        (KeyCode::Down, Key::Down),
        (KeyCode::Left, Key::Left),
        (KeyCode::Right, Key::Right),
        (KeyCode::PageUp, Key::PageUp),
        (KeyCode::PageDown, Key::PageDown),
    ];
    for (mq, ours) in map {
        if is_key_pressed(mq) {
            events.push(KeyEvent::new(ours, KeyModifiers::NONE));
        }
    }
    events
}

#[macroquad::main(window_conf)]
async fn main() {
    let cli = Cli::parse();
    let config = match Config::load(&cli) {
        Ok(config) => config,
        Err(err) => {
            eprintln!("{}: {err}", messages::CONFIG_ERROR);
            return;
        }
    };
    let conn = match Connection::connect(&config.server) {
        Ok(conn) => conn,
        Err(err) => {
            eprintln!("{}: {err}", config.server);
            return;
        }
    };
    let (reader, mut outgoing) = conn.split();
    for line in config.startup_lines() {
        let _ = outgoing.send_line(&line);
    }

    // The server -> a channel, in a thread of its own (just like lesson 06).
    let (tx, rx) = mpsc::channel();
    thread::spawn(move || {
        for line in reader.lines() {
            let Ok(line) = line else { break };
            if tx.send(parse_line(&line)).is_err() {
                return;
            }
        }
    });

    let atlas = Texture2D::from_file_with_format(include_bytes!("../assets/tiles.png"), None);
    atlas.set_filter(FilterMode::Nearest); // crisp pixels, no blur

    let mut app = App::new();
    app.push_log(Kind::Info, &format!("{} {}", messages::CONNECTED_TO, config.server));
    loop {
        for msg in rx.try_iter() {
            app.on_server(msg);
        }
        for key in keys() {
            if let Some(line) = app.on_key(key) {
                if outgoing.send_line(&line).is_err() {
                    app.on_disconnect();
                }
            }
        }
        if app.quit {
            break;
        }

        clear_background(BLACK);
        let mut renderer = GfxRenderer { atlas: &atlas, scale: 2.0 };
        let size = renderer.cell_px();
        let (cols, rows) = ((screen_width() / size) as u16, (screen_height() / size) as u16);
        draw_scene(&build_scene(&app, cols, rows), &mut renderer);
        next_frame().await;
    }
}
