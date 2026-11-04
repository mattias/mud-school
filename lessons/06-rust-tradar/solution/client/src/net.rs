//! Nätverksdelen av klienten.

use std::io::{self, BufRead, BufReader, Write};
use std::net::TcpStream;

/// En anslutning till MUD-servern.
pub struct Connection {
    reader: BufReader<TcpStream>,
    outgoing: Outgoing,
}

/// Skrivhalvan av en anslutning.
pub struct Outgoing {
    stream: TcpStream,
}

impl Outgoing {
    /// Skickar en rad till servern (lägger till "\r\n").
    pub fn send_line(&mut self, line: &str) -> io::Result<()> {
        self.stream.write_all(line.as_bytes())?;
        self.stream.write_all(b"\r\n")?;
        self.stream.flush()
    }
}

impl Connection {
    /// Ansluter till servern på `addr`, t.ex. "127.0.0.1:4000".
    pub fn connect(addr: &str) -> io::Result<Connection> {
        let stream = TcpStream::connect(addr)?;
        let writer = stream.try_clone()?;
        Ok(Connection {
            reader: BufReader::new(stream),
            outgoing: Outgoing { stream: writer },
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
        self.outgoing.send_line(line)
    }

    /// Delar upp anslutningen i en läs- och en skrivhalva som kan användas
    /// från olika trådar. `self` tas som värde: anslutningen *förbrukas*.
    pub fn split(self) -> (BufReader<TcpStream>, Outgoing) {
        (self.reader, self.outgoing)
    }
}
