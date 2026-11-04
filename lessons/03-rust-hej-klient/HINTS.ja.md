# ヒント

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

次に `\r\n` を取り除きます。`str::trim_end_matches` を見てください。文字の
配列を渡せます: `buf.trim_end_matches(['\r', '\n'])`。戻り値は `&str` なので、
`.to_string()` で `String` にします。

## send_line

`write_all` はバイト列（`&[u8]`）を取ります。`&str` は `.as_bytes()` で
バイト列になり、バイト文字列は `b"\r\n"` と書きます。最後に `self.writer.flush()`。

## main.rs

```rust
use std::io::{self, BufRead};
use mud_client::messages;
use mud_client::net::Connection;   // ライブラリ名は mud_client（Cargo.toml より）

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
