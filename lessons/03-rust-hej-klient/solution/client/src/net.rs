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
        let stream = TcpStream::connect(addr)?;
        let writer = stream.try_clone()?;
        Ok(Connection {
            reader: BufReader::new(stream),
            writer,
        })
    }

    /// Läser nästa rad från servern utan "\r\n" i slutet.
    /// Returnerar Ok(None) när servern har stängt anslutningen.
    pub fn read_line(&mut self) -> io::Result<Option<String>> {
        let mut buf = String::new();
        if self.reader.read_line(&mut buf)? == 0 {
            return Ok(None);
        }
        let line = buf.trim_end_matches(['\r', '\n']);
        Ok(Some(line.to_string()))
    }

    /// Skickar en rad till servern (lägger till "\r\n").
    pub fn send_line(&mut self, line: &str) -> io::Result<()> {
        self.writer.write_all(line.as_bytes())?;
        self.writer.write_all(b"\r\n")?;
        self.writer.flush()
    }
}
