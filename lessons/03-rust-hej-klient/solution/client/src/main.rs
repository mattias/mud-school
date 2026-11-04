// Startpunkten för din MUD-klient. Kör med: ./mudschool run client [adress]

use std::io::{self, BufRead};

use mud_client::net::Connection;

fn main() -> io::Result<()> {
    let addr = std::env::args()
        .nth(1)
        .unwrap_or_else(|| "127.0.0.1:4000".to_string());
    let mut conn = Connection::connect(&addr)?;
    println!("Ansluten till {addr}");

    // Välkomstmeddelandet
    if let Some(line) = conn.read_line()? {
        println!("{line}");
    }

    // Enkel turordning: läs en rad från tangentbordet, skicka den,
    // skriv ut svaret. (I lektion 6 gör vi detta samtidigt med trådar.)
    for input in io::stdin().lock().lines() {
        conn.send_line(&input?)?;
        match conn.read_line()? {
            Some(reply) => println!("{reply}"),
            None => {
                println!("Servern stängde anslutningen.");
                break;
            }
        }
    }
    Ok(())
}
