// Startpunkten för din MUD-klient. Kör med: ./mudschool run client [adress]

use std::io::{self, BufRead};
use std::thread;

use mud_client::net::Connection;
use mud_client::reader::spawn_reader;

fn main() -> io::Result<()> {
    let addr = std::env::args()
        .nth(1)
        .unwrap_or_else(|| "127.0.0.1:4000".to_string());
    let conn = Connection::connect(&addr)?;
    println!("Ansluten till {addr}");

    let (reader, mut outgoing) = conn.split();
    let lines = spawn_reader(reader);

    // Tråd som skriver ut allt servern skickar, så fort det kommer.
    thread::spawn(move || {
        for line in lines {
            println!("{line}");
        }
        println!("Servern stängde anslutningen.");
        std::process::exit(0);
    });

    // Huvudtråden läser från tangentbordet.
    for input in io::stdin().lock().lines() {
        outgoing.send_line(&input?)?;
    }
    Ok(())
}
