//! The network part of the client.

use std::io::{self, BufRead, BufReader, Write};
use std::net::TcpStream;

/// A connection to the MUD server.
pub struct Connection {
    reader: BufReader<TcpStream>,
    outgoing: Outgoing,
}

/// The writing half of a connection.
pub struct Outgoing {
    stream: TcpStream,
}

impl Outgoing {
    /// Sends a line to the server (appends "\r\n").
    pub fn send_line(&mut self, line: &str) -> io::Result<()> {
        self.stream.write_all(line.as_bytes())?;
        self.stream.write_all(b"\r\n")?;
        self.stream.flush()
    }
}

impl Connection {
    /// Connects to the server at `addr`, e.g. "127.0.0.1:4000".
    pub fn connect(addr: &str) -> io::Result<Connection> {
        let stream = TcpStream::connect(addr)?;
        let writer = stream.try_clone()?;
        Ok(Connection {
            reader: BufReader::new(stream),
            outgoing: Outgoing { stream: writer },
        })
    }

    /// Reads the next line from the server without the trailing "\r\n".
    /// Returns Ok(None) when the server has closed the connection.
    pub fn read_line(&mut self) -> io::Result<Option<String>> {
        let mut buf = String::new();
        if self.reader.read_line(&mut buf)? == 0 {
            return Ok(None);
        }
        let line = buf.trim_end_matches(['\r', '\n']);
        Ok(Some(line.to_string()))
    }

    /// Sends a line to the server (appends "\r\n").
    pub fn send_line(&mut self, line: &str) -> io::Result<()> {
        self.outgoing.send_line(line)
    }

    /// Splits the connection into a reading and a writing half that can be used
    /// from different threads. `self` is taken by value: the connection is *consumed*.
    pub fn split(self) -> (BufReader<TcpStream>, Outgoing) {
        (self.reader, self.outgoing)
    }
}
