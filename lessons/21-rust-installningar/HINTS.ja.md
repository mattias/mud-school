# ヒント

## Option で優先順位

```rust
server: cli.server.clone()
    .or_else(|| file.server.clone())
    .unwrap_or_else(|| DEFAULT_SERVER.to_string()),
name: cli.name.clone().or_else(|| file.name.clone()),
```

## startup_lines

`Option` は要素が 0 個か 1 個のイテレーターです:

```rust
let mut lines = vec![HELLO.to_string()];
lines.extend(self.name.clone());
lines
```

## secret 中の Enter

`KeyCode::Enter` のアームの先頭に置きます:

```rust
let line = std::mem::take(&mut self.input);
if self.secret {
    self.secret = false;
    return Some(line);
}
```

## Up/Down

match ガードで履歴をパスワードから遠ざけます:

```rust
KeyCode::Up if !self.secret => self.history_up(),
```

## ExitCode を返す main

```rust
fn main() -> ExitCode {
    let cli = Cli::parse();
    let config = match Config::load(&cli) {
        Ok(config) => config,
        Err(err) => {
            eprintln!("{}: {err}", messages::CONFIG_ERROR);
            return ExitCode::FAILURE;
        }
    };
    // ... the old main, in a function run(&config) -> io::Result<()> ...
}
```
