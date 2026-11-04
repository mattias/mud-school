# Ledtrådar

## Trådens kropp

```rust
thread::spawn(move || {
    for line in reader.lines() {
        // line är io::Result<String>
    }
});
```

`lines()` tar redan bort både `\n` och `\r\n`.

## Fel och stängda kanaler

`let Ok(line) = line else { break };` avbryter loopen vid läsfel.
`tx.send(line)` ger `Err` om mottagaren har släppts – då kan tråden också sluta.

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

(Om du byter fältet `writer` mot `outgoing: Outgoing` blir det `(self.reader, self.outgoing)`.)
