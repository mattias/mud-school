# Ledtrådar

## Fälten

Titta i `docs/sv/PROTOCOL.md` eller i Go-serverns `frame.go` – namnen ska vara
exakt desamma som i JSON.

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

`legend` har `String`-nycklar men vi frågar med en `char`:

```rust
self.legend.get(code.to_string().as_str()).map(String::as_str)
```

`.map(String::as_str)` gör om `Option<&String>` till `Option<&str>`.

## parse_line

```rust
if line.starts_with('{') {
    if let Ok(msg) = serde_json::from_str(line) {
        return msg;
    }
}
ServerMsg::Text { text: line.to_string() }
```

## Kompilerar inte?

`#[derive(Default)]` på en struct kräver att alla fält har `Default` – det
har `String`, `usize`, `Vec` och `HashMap`.
