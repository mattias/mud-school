//! The network part of the client.

use std::io::{self, BufRead, BufReader, Write};
use std::net::TcpStream;

/// A connection to the MUD server.
pub struct Connection {
    reader: BufReader<TcpStream>,
    writer: TcpStream,
}

impl Connection {
    /// Connects to the server at `addr`, e.g. "127.0.0.1:4000".
    pub fn connect(addr: &str) -> io::Result<Connection> {
        let stream = TcpStream::connect(addr)?;
        let writer = stream.try_clone()?;
        Ok(Connection {
            reader: BufReader::new(stream),
            writer,
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
        self.writer.write_all(line.as_bytes())?;
        self.writer.write_all(b"\r\n")?;
        self.writer.flush()
    }
}
