# Hints

## Precedence with Option

```rust
server: cli.server.clone()
    .or_else(|| file.server.clone())
    .unwrap_or_else(|| DEFAULT_SERVER.to_string()),
name: cli.name.clone().or_else(|| file.name.clone()),
```

## startup_lines

An `Option` is an iterator with zero or one elements:

```rust
let mut lines = vec![HELLO.to_string()];
lines.extend(self.name.clone());
lines
```

## Enter while secret

Put this first in the `KeyCode::Enter` arm:

```rust
let line = std::mem::take(&mut self.input);
if self.secret {
    self.secret = false;
    return Some(line);
}
```

## Up/Down

A match guard keeps the history away from passwords:

```rust
KeyCode::Up if !self.secret => self.history_up(),
```

## main with ExitCode

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
