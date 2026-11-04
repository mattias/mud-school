# Hints

## The thread's body

```rust
thread::spawn(move || {
    for line in reader.lines() {
        // line is io::Result<String>
    }
});
```

`lines()` already removes both `\n` and `\r\n`.

## Errors and closed channels

`let Ok(line) = line else { break };` stops the loop on a read error.
`tx.send(line)` returns `Err` if the receiver has been dropped – then the
thread can stop too.

## split

```rust
pub struct Outgoing {
    stream: TcpStream,
}

impl Connection {
    pub fn split(self) -> (BufReader<TcpStream>, Outgoing) {
        (self.reader, Outgoing { stream: self.writer })
    }
}
```

(If you replace the field `writer` with `outgoing: Outgoing` it becomes `(self.reader, self.outgoing)`.)
