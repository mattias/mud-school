# ヒント

## フィールド

`docs/ja/PROTOCOL.md` か Go サーバーの `frame.go` を見てください。名前は JSON と
まったく同じでなければなりません。

```rust
pub struct MapData {
    pub width: usize,
    pub height: usize,
    pub rows: Vec<String>,
    #[serde(default)]
    pub legend: HashMap<String, String>,
}
```

## meaning

`legend` のキーは `String` ですが、問い合わせは `char` で行います:

```rust
self.legend.get(code.to_string().as_str()).map(String::as_str)
```

`.map(String::as_str)` で `Option<&String>` を `Option<&str>` に変えます。

## parse_line

```rust
if line.starts_with('{') {
    if let Ok(msg) = serde_json::from_str(line) {
        return msg;
    }
}
ServerMsg::Text { text: line.to_string() }
```

## コンパイルできない？

構造体に `#[derive(Default)]` を付けるには、すべてのフィールドが `Default` を
持っている必要があります。`String`、`usize`、`Vec`、`HashMap` はどれも持っています。
