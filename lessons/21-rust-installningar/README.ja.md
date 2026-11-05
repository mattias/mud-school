# レッスン 21 – 設定とログイン

レッスン 20 からサーバーはパスワードを尋ねます。ところが今のクライアントはそれを平文で表示して
しまい、肩越しにのぞく誰にでも見えてしまいます。それに、クライアントを起動するたびに
アドレスと名前を入力し直しています。

このレッスンでクライアントに加わるもの:

- **コマンドラインフラグ**: `mud-client --name Anna 10.0.0.5:4000`。自分で書いていない `--help` 付き。
- **設定ファイル** `mudclient.toml`。
- 何がどこで失敗したかを伝える独自の **エラー型**。
- **隠されたパスワード入力欄**。

## まず: 依存関係

`workspace/client/Cargo.toml` の `[dependencies]` に:

```toml
clap = { version = "4", features = ["derive"] }
thiserror = "2"
toml = "0.8"
```

（または `cargo add clap --features derive` と `cargo add thiserror toml`）。

## 目標

- 他のクレートの **derive マクロ** を使う: `#[derive(Parser)]` が引数パーサーをまるごと書く。
- serde で TOML を読む: `#[serde(default, deny_unknown_fields)]`。
- `thiserror` で **エラーの enum** を書く: `#[error("...")]` と `#[source]`。
- 「見つからない」を他の I/O エラーと区別して扱う（`io::ErrorKind::NotFound`）。
- `Option` を組み合わせる: `a.or(b).unwrap_or(c)` で優先順位を 1 行で。
- `main` から `ExitCode` を返す。

## Rust の概念

### clap

```rust
#[derive(Debug, Clone, Default, PartialEq, Parser)]
#[command(name = "mud-client", version, about = "The MUD School client")]
pub struct Cli {
    /// The server to connect to, e.g. 127.0.0.1:4000
    pub server: Option<String>,          // 位置引数
    /// Log in with this name automatically
    #[arg(long)]
    pub name: Option<String>,            // --name Anna
}
```

`///` コメントがヘルプのテキストになります。`Cli::parse()` は `std::env::args()` を読み、
エラーなら分かりやすいメッセージを出して終了します。`Cli::try_parse_from([...])` は
`Result` を返します。テストが使うのはこちらです。

### thiserror

エラー型は普通の enum です。`Display` と `std::error::Error` は `thiserror` が書いてくれます:

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

`#[source]` は元のエラーを保持するので、受け取った側がさらに掘り下げられます
（`err.source()`）。Go の `%w` や `errors.Unwrap` と比べてみましょう。

### 存在しないファイル

設定ファイルがないのはまったく問題ありません。「既定値を使う」という意味です。
読む権限のないファイルや、誤りのあるファイルは問題です:

```rust
let text = match fs::read_to_string(path) {
    Ok(text) => text,
    Err(e) if e.kind() == io::ErrorKind::NotFound => return Ok(FileConfig::default()),
    Err(source) => return Err(ConfigError::Read { path: path.to_path_buf(), source }),
};
```

## 契約

新しいモジュール `config.rs`（`lib.rs` に `pub mod config;`）:

- `pub const DEFAULT_SERVER: &str = "127.0.0.1:4000"` と `DEFAULT_CONFIG = "mudclient.toml"`。
- `Cli`（clap。`Debug, Clone, Default, PartialEq, Parser` を derive）: `server`（位置引数）、`--name`、`--config`（`PathBuf`）。すべて `Option`。
- `FileConfig`（`Debug, Clone, Default, PartialEq, Deserialize`、`#[serde(default, deny_unknown_fields)]`）: `server`、`name`。どちらも `Option<String>`。
  - `FileConfig::parse(&str) -> Result<FileConfig, toml::de::Error>`
  - `FileConfig::load(&Path) -> Result<FileConfig, ConfigError>` – ファイルがなければ `FileConfig::default()`。
- `ConfigError`（thiserror）: `Read { path, source: io::Error }` と `Parse { path, source: toml::de::Error }`。メッセージはパスで始まる。
- `Config { server: String, name: Option<String> }`（`Debug, Clone, PartialEq`）:
  - `Config::merge(&Cli, &FileConfig)` – コマンドライン、ファイル、既定値の順に優先。
  - `Config::load(&Cli)` – `--config`（または `DEFAULT_CONFIG`）を読んでマージする。
  - `startup_lines(&self) -> Vec<String>` – `HELLO`、そして名前があれば名前。

プロトコルとアプリ:

- `{"type":"prompt","prompt":"password"}` 用の `ServerMsg::Prompt { prompt: String }`（`#[serde(default)]`）。
- `App::secret: bool` – prompt が `"password"` か `"new_password"` なら true、それ以外の prompt なら false。
- `secret` の間: 入力欄は 1 文字につき `*` を 1 つ表示し、枠の見出しは
  `messages::TITLE_PASSWORD`、上下キーは何もせず、Enter は行をログにも履歴にも **入れずに**
  返す。その後 `secret` は再び false。

`main.rs`: `Cli::parse()`、`Config::load`（エラーは `messages::CONFIG_ERROR` とともに表示し、
`ExitCode::FAILURE` を返す）、そして `HELLO` だけを送る代わりに `startup_lines()`。

## 課題

`config.rs` を書き、次に `protocol.rs` と `app.rs` の prompt、最後に `main.rs`。試してみましょう:

```
./mudschool run server
./mudschool run client --help
./mudschool run client --name Anna
```

そして `workspace/client/mudclient.toml` というファイルで:

```toml
server = "127.0.0.1:4000"
name = "Anna"
```

## 考えてみよう

- 綴りを間違えたフィールド（`sever = ...`）を無視せずエラーにすべきなのはなぜ？
- パスワードは画面では隠れていますが、サーバーまではどうやって届く？ ネットワーク上で守るには何が必要？
- パスワードを `mudclient.toml` に書けるようにすべき？ どんな危険がある？
- `ConfigError` にはバリアントが 2 つあります。多くのバリアントが役立つのはどんなときで、`String` 1 つで十分なのはどんなとき？

## チャレンジ

- ユーザーの設定フォルダーでも設定ファイルを探す（`dirs` クレート。Linux なら `~/.config/mudschool/`）。
- `--log <ファイル>`: ログの内容をファイルにも書き出す。
- 色の設定: `mudclient.toml` で `style_for` を上書きできるようにする。
