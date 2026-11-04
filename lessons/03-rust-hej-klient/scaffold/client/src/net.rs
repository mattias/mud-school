//! Nätverksdelen av klienten.

use std::io::{self, BufRead, BufReader, Write};
use std::net::TcpStream;

/// En anslutning till MUD-servern.
pub struct Connection {
    reader: BufReader<TcpStream>,
    writer: TcpStream,
}

impl Connection {
    /// Ansluter till servern på `addr`, t.ex. "127.0.0.1:4000".
    pub fn connect(addr: &str) -> io::Result<Connection> {
        // TODO(lektion 3): anslut med TcpStream::connect(addr)?
        // Du behöver TVÅ handtag till samma ström: ett att läsa från och ett
        // att skriva till. Tips: stream.try_clone()?
        todo!()
    }

    /// Läser nästa rad från servern utan "\r\n" i slutet.
    /// Returnerar Ok(None) när servern har stängt anslutningen.
    pub fn read_line(&mut self) -> io::Result<Option<String>> {
        // TODO(lektion 3): använd self.reader.read_line(&mut buf).
        // Den returnerar antal lästa bytes – 0 betyder att strömmen är slut.
        todo!()
    }

    /// Skickar en rad till servern (lägger till "\r\n").
    pub fn send_line(&mut self, line: &str) -> io::Result<()> {
        // TODO(lektion 3): skriv line och "\r\n" till self.writer.
        todo!()
    }
}
