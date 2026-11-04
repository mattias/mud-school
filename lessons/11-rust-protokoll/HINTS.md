# Hints

## The fields

Look in `docs/en/PROTOCOL.md` or in the Go server's `frame.go` – the names
must be exactly the same as in the JSON.

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

`legend` has `String` keys but we ask with a `char`:

```rust
self.legend.get(code.to_string().as_str()).map(String::as_str)
```

`.map(String::as_str)` turns `Option<&String>` into `Option<&str>`.

## parse_line

```rust
if line.starts_with('{') {
    if let Ok(msg) = serde_json::from_str(line) {
        return msg;
    }
}
ServerMsg::Text { text: line.to_string() }
```

## Doesn't compile?

`#[derive(Default)]` on a struct requires that every field has `Default` –
`String`, `usize`, `Vec` and `HashMap` all do.
