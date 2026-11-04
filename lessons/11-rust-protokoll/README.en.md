# Lesson 11 – The protocol in Rust

The server can speak JSON now. Time for the client to understand it. In Go we
sent a flat struct `Frame` with a `Type` field. In Rust we do something much
more powerful: an **enum in which every variant carries its own data**.

```rust
pub enum ServerMsg {
    Hello { text: String },
    Text { text: String },
    Map { map: MapData },
    Status { status: Status },
    Unknown,
}
```

A `ServerMsg::Map` *always* has a map, and a `ServerMsg::Text` *never* has
one. Invalid states cannot even be expressed. Compare with Go, where
`Frame{Type: "text", Map: &m}` compiles without complaint.

## Goals

- Use enums with data and `match` to handle every case.
- Add external **crates** with Cargo.
- Use `serde` to decode JSON with attributes like `tag`, `default`, `other`.
- Implement a trait from the standard library: `Display`.
- Think about **forward compatibility**: an old client must not crash against a new server.

## Rust concepts

### match must be exhaustive

```rust
match msg {
    ServerMsg::Text { text } => println!("{text}"),
    ServerMsg::Map { map } => draw(&map),
    ServerMsg::Status { status } => println!("[{status}]"),
    ServerMsg::Hello { .. } | ServerMsg::Unknown => {}   // .. = don't care about the fields
}
```

If you add a variant and forget a place, the compiler refuses. That is one of
Rust's greatest strengths as a program grows.

`let ... else` is handy when you only care about *one* case:

```rust
let ServerMsg::Map { map } = msg else {
    return;   // must leave the function
};
```

### Crates and serde

```
cd workspace/client
cargo add serde --features derive
cargo add serde_json
```

```rust
#[derive(Deserialize)]
#[serde(tag = "type", rename_all = "lowercase")]  // {"type":"text", ...} -> Text
enum ServerMsg {
    Text {
        #[serde(default)]   // field missing? use String::default() = ""
        text: String,
    },
    #[serde(other)]         // every unknown "type" ends up here
    Unknown,
}

let msg: ServerMsg = serde_json::from_str(line)?;
```

**Go ↔ Rust:** the struct tags `` `json:"max_hp"` `` in Go correspond to
attributes like `#[serde(rename = "max_hp")]` in Rust. They are not needed
here – the fields already have the same names.

### Traits: Display

```rust
impl fmt::Display for Status {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{} · HP {}/{}", self.name, self.hp, self.max_hp)
    }
}
// Now println!("{status}") and status.to_string() work
```

**Go ↔ Rust:** the same idea as the `String() string` method (`fmt.Stringer`)
in Go – but in Rust you implement the trait *explicitly*.

## Task

1. Add `serde` and `serde_json` (see above).
2. **`src/protocol.rs`** (new file): fill in the variants, the structs,
   `MapData::meaning`, `Display` for `Status` (use `messages::ONLINE` for the
   word "online"), and `parse_line`:
   - lines that start with `{` and are valid JSON become the right variant,
   - everything else (plain text, broken JSON) becomes `ServerMsg::Text` with the whole line.
3. Add `pub mod protocol;` to `lib.rs`.
4. **`src/main.rs`**: send `HELLO` right after connecting and print the
   messages nicely: text as it is, the map's rows, the status inside `[...]`.

## Try it yourself

```
./mudschool run server
./mudschool run client
```

Walk around the world – now the map is drawn after every step!

## Think about

- Why is it better for an unknown frame type to become `Unknown` than an error?
- `parse_line` returns `ServerMsg`, not `Result<ServerMsg, _>`. What do we
  gain and lose by that?

## Challenges

- Write your own error type `enum ProtocolError { NotJson, BadFrame(serde_json::Error) }`
  and a `fn parse_frame(line) -> Result<ServerMsg, ProtocolError>`. Implement
  `Display` and `std::error::Error` for it.
- Add `Serialize` and write a test that encodes and decodes (a round trip).
