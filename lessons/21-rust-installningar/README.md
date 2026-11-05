# Lesson 21 – Settings and login

Since lesson 20 the server asks for a password – and right now your client
shows it in plain text, for anyone looking over your shoulder. And every time
you start the client you type the address and your name again.

In this lesson the client gets:

- **command-line flags**: `mud-client --name Anna 10.0.0.5:4000`, with a `--help` you never wrote;
- a **settings file** `mudclient.toml`;
- its own **error types**, which say what went wrong and where;
- a **hidden password field**.

## First: the dependencies

In `workspace/client/Cargo.toml`, under `[dependencies]`:

```toml
clap = { version = "4", features = ["derive"] }
thiserror = "2"
toml = "0.8"
```

(or `cargo add clap --features derive` and `cargo add thiserror toml`).

## Goals

- Use **derive macros** from other crates: `#[derive(Parser)]` writes a whole argument parser.
- Read TOML with serde: `#[serde(default, deny_unknown_fields)]`.
- Write an **error enum** with `thiserror`: `#[error("...")]` and `#[source]`.
- Handle "not found" differently from other I/O errors (`io::ErrorKind::NotFound`).
- Combine `Option`s: `a.or(b).unwrap_or(c)` – precedence in one line.
- Return an `ExitCode` from `main`.

## Rust concepts

### clap

```rust
#[derive(Debug, Clone, Default, PartialEq, Parser)]
#[command(name = "mud-client", version, about = "The MUD School client")]
pub struct Cli {
    /// The server to connect to, e.g. 127.0.0.1:4000
    pub server: Option<String>,          // a positional argument
    /// Log in with this name automatically
    #[arg(long)]
    pub name: Option<String>,            // --name Anna
}
```

The `///` comments become the help text. `Cli::parse()` reads
`std::env::args()` and exits with a nice message on errors;
`Cli::try_parse_from([...])` returns a `Result` – that's what the tests use.

### thiserror

An error type is an ordinary enum. `thiserror` writes `Display` and
`std::error::Error` for you:

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

`#[source]` keeps the original error, so that whoever catches it can dig
deeper (`err.source()`). Compare with Go's `%w` and `errors.Unwrap`.

### A file that doesn't exist

No settings file is fine – it means "use the defaults". A file you may not
read, or one with errors, is not fine:

```rust
let text = match fs::read_to_string(path) {
    Ok(text) => text,
    Err(e) if e.kind() == io::ErrorKind::NotFound => return Ok(FileConfig::default()),
    Err(source) => return Err(ConfigError::Read { path: path.to_path_buf(), source }),
};
```

## The contract

A new module `config.rs` (`pub mod config;` in `lib.rs`):

- `pub const DEFAULT_SERVER: &str = "127.0.0.1:4000"` and `DEFAULT_CONFIG = "mudclient.toml"`.
- `Cli` (clap, derives `Debug, Clone, Default, PartialEq, Parser`): `server` (positional), `--name`, `--config` (a `PathBuf`) – all `Option`.
- `FileConfig` (`Debug, Clone, Default, PartialEq, Deserialize`, `#[serde(default, deny_unknown_fields)]`): `server`, `name` – both `Option<String>`.
  - `FileConfig::parse(&str) -> Result<FileConfig, toml::de::Error>`
  - `FileConfig::load(&Path) -> Result<FileConfig, ConfigError>` – a missing file gives `FileConfig::default()`.
- `ConfigError` (thiserror): `Read { path, source: io::Error }` and `Parse { path, source: toml::de::Error }`; the message starts with the path.
- `Config { server: String, name: Option<String> }` (`Debug, Clone, PartialEq`):
  - `Config::merge(&Cli, &FileConfig)` – command line before file before defaults.
  - `Config::load(&Cli)` – reads `--config` (or `DEFAULT_CONFIG`) and merges.
  - `startup_lines(&self) -> Vec<String>` – `HELLO`, then the name if there is one.

The protocol and the app:

- `ServerMsg::Prompt { prompt: String }` (`#[serde(default)]`) for `{"type":"prompt","prompt":"password"}`.
- `App::secret: bool` – true after a prompt `"password"` or `"new_password"`, false after any other prompt.
- While `secret`: the input shows one `*` per character, the box title is
  `messages::TITLE_PASSWORD`, Up/Down do nothing, and Enter returns the line
  **without** putting it in the log or the history – then `secret` is false again.

`main.rs`: `Cli::parse()`, `Config::load` (an error is printed with
`messages::CONFIG_ERROR` and gives `ExitCode::FAILURE`), and
`startup_lines()` instead of just sending `HELLO`.

## Task

Write `config.rs`, then the prompt in `protocol.rs` and `app.rs`, then
`main.rs`. Try it:

```
./mudschool run server
./mudschool run client --help
./mudschool run client --name Anna
```

and with a file `workspace/client/mudclient.toml`:

```toml
server = "127.0.0.1:4000"
name = "Anna"
```

## Think about

- Why should a misspelt field (`sever = ...`) be an error instead of being ignored?
- The password is hidden on screen, but how does it travel to the server? What would be needed to protect it on the network?
- Should the password be allowed in `mudclient.toml`? What are the risks?
- `ConfigError` has two variants. When is it worth having many variants, and when is one `String` enough?

## Challenges

- Look for the settings file in the user's config folder too (the `dirs` crate: `~/.config/mudschool/` on Linux).
- `--log <file>`: write everything in the log to a file as well.
- Settings for colours: let `mudclient.toml` override `style_for`.
