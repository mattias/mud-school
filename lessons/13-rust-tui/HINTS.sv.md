# Ledtrådar

## push_log

```rust
self.log.extend(text.lines().map(str::to_string));
if self.log.len() > MAX_LOG {
    let extra = self.log.len() - MAX_LOG;
    self.log.drain(..extra);
}
```

## on_key

```rust
if key.kind != KeyEventKind::Press {
    return None;
}
match key.code {
    KeyCode::Esc => self.quit = true,
    KeyCode::Char('c') if key.modifiers.contains(KeyModifiers::CONTROL) => self.quit = true,
    KeyCode::Char(c) => self.input.push(c),
    KeyCode::Backspace => { self.input.pop(); }
    KeyCode::Enter => return Some(std::mem::take(&mut self.input)),
    _ => {}
}
None
```

Ordningen på `match`-grenarna spelar roll: `Char('c') if ...` måste stå före `Char(c)`.

## draw_log

Ramen tar två rader. Visa bara de sista `area.height - 2` raderna:

```rust
let visible = area.height.saturating_sub(2) as usize;
let start = app.log.len().saturating_sub(visible);
let lines: Vec<Line> = app.log[start..].iter().map(|l| Line::from(l.as_str())).collect();
```

## draw_map

```rust
render_map(map)
    .into_iter()
    .map(|row| {
        row.into_iter()
            .map(|c| Span::styled(c.glyph.to_string(), Style::default().fg(c.fg.into())))
            .collect::<Line>()
    })
    .collect::<Vec<Line>>()
```

## Huvudloopen

```rust
let mut terminal = ratatui::init();
let mut app = App::new();
while !app.quit {
    terminal.draw(|f| draw(f, &app))?;
    match rx.recv() {
        Ok(Event::Key(key)) => { if let Some(line) = app.on_key(key) { outgoing.send_line(&line)?; } }
        Ok(Event::Server(msg)) => app.on_server(msg),
        Ok(Event::Disconnected) => app.on_disconnect(),
        Err(_) => break,
    }
}
ratatui::restore();
```

Se till att `ratatui::restore()` körs även om `?` returnerar tidigt – t.ex.
genom att lägga loopen i en egen funktion och anropa `restore` efteråt.
