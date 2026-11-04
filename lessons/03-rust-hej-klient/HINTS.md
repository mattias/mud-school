# Hints

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

Then you want to remove `\r\n`. Look at `str::trim_end_matches` – it can take
an array of characters: `buf.trim_end_matches(['\r', '\n'])`. It returns a
`&str`; turn it into a `String` with `.to_string()`.

## send_line

`write_all` takes bytes (`&[u8]`). A `&str` becomes bytes with `.as_bytes()`,
and a byte string is written `b"\r\n"`. Finish with `self.writer.flush()`.

## main.rs

```rust
use std::io::{self, BufRead};
use mud_client::messages;
use mud_client::net::Connection;   // your library is called mud_client (from Cargo.toml)

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
