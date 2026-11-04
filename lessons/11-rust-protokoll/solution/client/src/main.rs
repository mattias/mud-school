// Entry point of your MUD client. Run it with: ./mudschool run client [address]

use std::io::{self, BufRead};
use std::thread;

use mud_client::messages;
use mud_client::net::Connection;
use mud_client::protocol::{parse_line, ServerMsg, HELLO};
use mud_client::reader::spawn_reader;

fn main() -> io::Result<()> {
    let addr = std::env::args()
        .nth(1)
        .unwrap_or_else(|| "127.0.0.1:4000".to_string());
    let conn = Connection::connect(&addr)?;
    println!("{} {addr}", messages::CONNECTED_TO);

    let (reader, mut outgoing) = conn.split();
    outgoing.send_line(HELLO)?; // ask for JSON
    let lines = spawn_reader(reader);

    thread::spawn(move || {
        for line in lines {
            match parse_line(&line) {
                ServerMsg::Text { text } => println!("{text}"),
                ServerMsg::Map { map } => {
                    for row in &map.rows {
                        println!("    {row}");
                    }
                }
                ServerMsg::Status { status } => println!("[{status}]"),
                ServerMsg::Hello { .. } | ServerMsg::Unknown => {}
            }
        }
        println!("{}", messages::SERVER_CLOSED);
        std::process::exit(0);
    });

    for input in io::stdin().lock().lines() {
        outgoing.send_line(&input?)?;
    }
    Ok(())
}
