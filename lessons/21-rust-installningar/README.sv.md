# Lektion 21 – Inställningar och inloggning

Sedan lektion 20 frågar servern efter ett lösenord – och just nu visar din
klient det i klartext, för vem som helst som tittar över axeln. Och varje gång
du startar klienten skriver du adressen och ditt namn igen.

I den här lektionen får klienten:

- **kommandoradsflaggor**: `mud-client --name Anna 10.0.0.5:4000`, med en `--help` du aldrig skrev;
- en **inställningsfil** `mudclient.toml`;
- egna **feltyper** som säger vad som gick fel och var;
- ett **dolt lösenordsfält**.

## Först: beroendena

I `workspace/client/Cargo.toml`, under `[dependencies]`:

```toml
clap = { version = "4", features = ["derive"] }
thiserror = "2"
toml = "0.8"
```

(eller `cargo add clap --features derive` och `cargo add thiserror toml`).

## Mål

- Använda **derive-makron** från andra crates: `#[derive(Parser)]` skriver en hel argumenttolk.
- Läsa TOML med serde: `#[serde(default, deny_unknown_fields)]`.
- Skriva en **fel-enum** med `thiserror`: `#[error("...")]` och `#[source]`.
- Hantera "finns inte" annorlunda än andra I/O-fel (`io::ErrorKind::NotFound`).
- Kombinera `Option`: `a.or(b).unwrap_or(c)` – prioritetsordning på en rad.
- Returnera en `ExitCode` från `main`.

## Rust-koncept

### clap

```rust
#[derive(Debug, Clone, Default, PartialEq, Parser)]
#[command(name = "mud-client", version, about = "The MUD School client")]
pub struct Cli {
    /// The server to connect to, e.g. 127.0.0.1:4000
    pub server: Option<String>,          // ett positionsargument
    /// Log in with this name automatically
    #[arg(long)]
    pub name: Option<String>,            // --name Anna
}
```

`///`-kommentarerna blir hjälptexten. `Cli::parse()` läser
`std::env::args()` och avslutar med ett snyggt meddelande vid fel;
`Cli::try_parse_from([...])` returnerar ett `Result` – det är vad testerna använder.

### thiserror

En feltyp är en vanlig enum. `thiserror` skriver `Display` och
`std::error::Error` åt dig:

```rust
#[derive(Debug, thiserror::Error)]
pub enum ConfigError {
    #[error("{}: {source}", path.display())]
    Parse {
        path: PathBuf,
        #[source]
        source: toml::de::Error,
    },
    // ...
}
```

`#[source]` behåller det ursprungliga felet, så att den som fångar det kan
gräva djupare (`err.source()`). Jämför med Go:s `%w` och `errors.Unwrap`.

### En fil som inte finns

Ingen inställningsfil är helt okej – det betyder "använd standardvärdena". En
fil du inte får läsa, eller en med fel i, är inte okej:

```rust
let text = match fs::read_to_string(path) {
    Ok(text) => text,
    Err(e) if e.kind() == io::ErrorKind::NotFound => return Ok(FileConfig::default()),
    Err(source) => return Err(ConfigError::Read { path: path.to_path_buf(), source }),
};
```

## Kontraktet

En ny modul `config.rs` (`pub mod config;` i `lib.rs`):

- `pub const DEFAULT_SERVER: &str = "127.0.0.1:4000"` och `DEFAULT_CONFIG = "mudclient.toml"`.
- `Cli` (clap, deriverar `Debug, Clone, Default, PartialEq, Parser`): `server` (positionsargument), `--name`, `--config` (en `PathBuf`) – alla `Option`.
- `FileConfig` (`Debug, Clone, Default, PartialEq, Deserialize`, `#[serde(default, deny_unknown_fields)]`): `server`, `name` – båda `Option<String>`.
  - `FileConfig::parse(&str) -> Result<FileConfig, toml::de::Error>`
  - `FileConfig::load(&Path) -> Result<FileConfig, ConfigError>` – en fil som saknas ger `FileConfig::default()`.
- `ConfigError` (thiserror): `Read { path, source: io::Error }` och `Parse { path, source: toml::de::Error }`; meddelandet börjar med sökvägen.
- `Config { server: String, name: Option<String> }` (`Debug, Clone, PartialEq`):
  - `Config::merge(&Cli, &FileConfig)` – kommandoraden före filen före standardvärdena.
  - `Config::load(&Cli)` – läser `--config` (eller `DEFAULT_CONFIG`) och slår ihop.
  - `startup_lines(&self) -> Vec<String>` – `HELLO`, sedan namnet om det finns.

Protokollet och appen:

- `ServerMsg::Prompt { prompt: String }` (`#[serde(default)]`) för `{"type":"prompt","prompt":"password"}`.
- `App::secret: bool` – true efter en prompt `"password"` eller `"new_password"`, false efter alla andra prompter.
- Medan `secret`: inmatningen visar en `*` per tecken, rutans rubrik är
  `messages::TITLE_PASSWORD`, Upp/Ner gör ingenting, och Enter returnerar
  raden **utan** att lägga den i loggen eller historiken – sedan är `secret` false igen.

`main.rs`: `Cli::parse()`, `Config::load` (ett fel skrivs ut med
`messages::CONFIG_ERROR` och ger `ExitCode::FAILURE`), och `startup_lines()`
i stället för att bara skicka `HELLO`.

## Uppgift

Skriv `config.rs`, sedan prompten i `protocol.rs` och `app.rs`, sedan
`main.rs`. Prova:

```
./mudschool run server
./mudschool run client --help
./mudschool run client --name Anna
```

och med en fil `workspace/client/mudclient.toml`:

```toml
server = "127.0.0.1:4000"
name = "Anna"
```

## Fundera på

- Varför ska ett felstavat fält (`sever = ...`) vara ett fel i stället för att ignoreras?
- Lösenordet är dolt på skärmen, men hur färdas det till servern? Vad skulle behövas för att skydda det på nätverket?
- Ska lösenordet få stå i `mudclient.toml`? Vilka är riskerna?
- `ConfigError` har två varianter. När lönar det sig med många varianter, och när räcker en `String`?

## Utmaningar

- Leta efter inställningsfilen i användarens konfigurationsmapp också (craten `dirs`: `~/.config/mudschool/` på Linux).
- `--log <fil>`: skriv allt i loggen till en fil också.
- Inställningar för färger: låt `mudclient.toml` ersätta `style_for`.
