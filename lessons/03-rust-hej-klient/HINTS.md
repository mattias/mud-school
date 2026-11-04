# Ledtrådar

## connect

```rust
let stream = TcpStream::connect(addr)?;
let writer = stream.try_clone()?;
Ok(Connection { reader: BufReader::new(stream), writer })
```

## read_line

```rust
let mut buf = String::new();
let n = self.reader.read_line(&mut buf)?;
if n == 0 {
    return Ok(None);
}
```

Sedan vill du ta bort `\r\n`. Titta på `str::trim_end_matches` – den kan ta
en array av tecken: `buf.trim_end_matches(['\r', '\n'])`. Den returnerar en
`&str`; gör om den till en `String` med `.to_string()`.

## send_line

`write_all` tar bytes (`&[u8]`). En `&str` blir bytes med `.as_bytes()`, och
en bytesträng skrivs `b"\r\n"`. Avsluta gärna med `self.writer.flush()`.

## main.rs

```rust
use std::io::{self, BufRead};
use mud_client::net::Connection;   // ditt bibliotek heter mud_client (från Cargo.toml)

fn main() -> io::Result<()> {
    let addr = std::env::args().nth(1).unwrap_or_else(|| "127.0.0.1:4000".to_string());
    let mut conn = Connection::connect(&addr)?;
    // ...
    for input in io::stdin().lock().lines() {
        let input = input?;
        // ...
    }
    Ok(())
}
```
