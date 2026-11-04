# Lektion 11 – Protokoll i Rust

Servern kan nu prata JSON. Dags för klienten att förstå den. I Go skickade
vi en platt struct `Frame` med ett `Type`-fält. I Rust gör vi något mycket
kraftfullare: en **enum där varje variant bär sin egen data**.

```rust
pub enum ServerMsg {
    Hello { text: String },
    Text { text: String },
    Map { map: MapData },
    Status { status: Status },
    Unknown,
}
```

Ett `ServerMsg::Map` *har* alltid en karta, och ett `ServerMsg::Text` har
*aldrig* en. Felaktiga tillstånd går inte ens att uttrycka. Jämför med Go,
där `Frame{Type: "text", Map: &m}` kompilerar utan protester.

## Mål

- Använda enums med data och `match` för att hantera alla fall.
- Lägga till externa **crates** med Cargo.
- Använda `serde` för att avkoda JSON med attribut som `tag`, `default`, `other`.
- Implementera en trait från standardbiblioteket: `Display`.
- Tänka på **framåtkompatibilitet**: en gammal klient ska inte krascha mot en ny server.

## Rust-koncept

### match måste vara uttömmande

```rust
match msg {
    ServerMsg::Text { text } => println!("{text}"),
    ServerMsg::Map { map } => draw(&map),
    ServerMsg::Status { status } => println!("[{status}]"),
    ServerMsg::Hello { .. } | ServerMsg::Unknown => {}   // .. = bry dig inte om fälten
}
```

Lägger du till en variant och glömmer ett ställe vägrar kompilatorn. Det
är en av Rusts största styrkor när ett program växer.

`let ... else` är praktiskt när du bara bryr dig om *ett* fall:

```rust
let ServerMsg::Map { map } = msg else {
    return;   // måste lämna funktionen
};
```

### Crates och serde

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
        #[serde(default)]   // saknas fältet? använd String::default() = ""
        text: String,
    },
    #[serde(other)]         // alla okända "type" hamnar här
    Unknown,
}

let msg: ServerMsg = serde_json::from_str(line)?;
```

**Go ↔ Rust:** struct-taggarna `` `json:"max_hp"` `` i Go motsvaras av
attribut som `#[serde(rename = "max_hp")]` i Rust. Här behövs de inte –
fälten heter redan likadant.

### Traits: Display

```rust
impl fmt::Display for Status {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{} · HP {}/{}", self.name, self.hp, self.max_hp)
    }
}
// Nu fungerar println!("{status}") och status.to_string()
```

**Go ↔ Rust:** samma idé som `String() string`-metoden (`fmt.Stringer`) i
Go – men i Rust implementerar man traiten *explicit*.

## Uppgift

1. Lägg till `serde` och `serde_json` (se ovan).
2. **`src/protocol.rs`** (ny fil): fyll i varianterna, structarna,
   `MapData::meaning`, `Display` för `Status`, och `parse_line`:
   - rader som börjar med `{` och är giltig JSON blir rätt variant,
   - allt annat (vanlig text, trasig JSON) blir `ServerMsg::Text` med hela raden.
3. Lägg till `pub mod protocol;` i `lib.rs`.
4. **`src/main.rs`**: skicka `HELLO` direkt efter anslutningen och skriv ut
   meddelandena snyggt: text som den är, kartans rader, statusen inom `[...]`.

## Testa själv

```
./mudschool run server
./mudschool run client
```

Gå runt i världen – nu ritas kartan efter varje steg!

## Fundera på

- Varför är det bättre att en okänd frame-typ blir `Unknown` än ett fel?
- `parse_line` returnerar `ServerMsg`, inte `Result<ServerMsg, _>`. Vad
  vinner och förlorar vi på det?

## Utmaningar

- Skriv en egen feltyp `enum ProtocolError { NotJson, BadFrame(serde_json::Error) }`
  och en `fn parse_frame(line) -> Result<ServerMsg, ProtocolError>`.
  Implementera `Display` och `std::error::Error` för den.
- Lägg till `Serialize` och skriv ett test som kodar och avkodar (round-trip).
