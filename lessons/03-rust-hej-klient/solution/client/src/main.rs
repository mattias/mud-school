// Entry point of your MUD client. Run it with: ./mudschool run client [address]

use std::io::{self, BufRead};

use mud_client::messages;
use mud_client::net::Connection;

fn main() -> io::Result<()> {
    let addr = std::env::args()
        .nth(1)
        .unwrap_or_else(|| "127.0.0.1:4000".to_string());
    let mut conn = Connection::connect(&addr)?;
    println!("{} {addr}", messages::CONNECTED_TO);

    // The welcome message
    if let Some(line) = conn.read_line()? {
        println!("{line}");
    }

    // Simple turn taking: read a line from the keyboard, send it, print the
    // reply. (In lesson 6 we do this concurrently with threads.)
    for input in io::stdin().lock().lines() {
        conn.send_line(&input?)?;
        match conn.read_line()? {
            Some(reply) => println!("{reply}"),
            None => {
                println!("{}", messages::SERVER_CLOSED);
                break;
            }
        }
    }
    Ok(())
}
