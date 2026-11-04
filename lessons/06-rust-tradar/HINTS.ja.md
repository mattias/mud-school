# ヒント

## スレッドの本体

```rust
thread::spawn(move || {
    for line in reader.lines() {
        // line は io::Result<String>
    }
});
```

`lines()` は `\n` も `\r\n` もすでに取り除いてくれます。

## エラーと閉じたチャネル

`let Ok(line) = line else { break };` で読み取りエラーのときにループを抜けます。
受信側が破棄されていると `tx.send(line)` は `Err` を返します。そのときもスレッドを
終えてかまいません。

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

（フィールド `writer` を `outgoing: Outgoing` に置き換えたなら `(self.reader, self.outgoing)` になります。）
