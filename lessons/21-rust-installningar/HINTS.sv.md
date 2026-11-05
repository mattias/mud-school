# Ledtrådar

## Prioritetsordning med Option

```rust
server: cli.server.clone()
    .or_else(|| file.server.clone())
    .unwrap_or_else(|| DEFAULT_SERVER.to_string()),
name: cli.name.clone().or_else(|| file.name.clone()),
```

## startup_lines

En `Option` är en iterator med noll eller ett element:

```rust
let mut lines = vec![HELLO.to_string()];
lines.extend(self.name.clone());
lines
```

## Enter medan secret

Lägg det här först i grenen `KeyCode::Enter`:

```rust
let line = std::mem::take(&mut self.input);
if self.secret {
    self.secret = false;
    return Some(line);
}
```

## Up/Down

Ett match-villkor (guard) håller historiken borta från lösenord:

```rust
KeyCode::Up if !self.secret => self.history_up(),
```

## main med ExitCode

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
