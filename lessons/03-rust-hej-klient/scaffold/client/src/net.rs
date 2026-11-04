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
        // TODO(lesson 3): connect with TcpStream::connect(addr)?
        // You need TWO handles to the same stream: one to read from and one
        // to write to. Hint: stream.try_clone()?
        todo!()
    }

    /// Reads the next line from the server without the trailing "\r\n".
    /// Returns Ok(None) when the server has closed the connection.
    pub fn read_line(&mut self) -> io::Result<Option<String>> {
        // TODO(lesson 3): use self.reader.read_line(&mut buf).
        // It returns the number of bytes read – 0 means the stream has ended.
        todo!()
    }

    /// Sends a line to the server (appends "\r\n").
    pub fn send_line(&mut self, line: &str) -> io::Result<()> {
        // TODO(lesson 3): write line and "\r\n" to self.writer.
        todo!()
    }
}
