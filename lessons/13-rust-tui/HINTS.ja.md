# ヒント

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

`match` の腕の順序は重要です: `Char('c') if ...` は `Char(c)` より前に置きます。

## draw_log

枠線で 2 行使います。最後の `area.height - 2` 行だけを表示しましょう:

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

## メインループ

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

`?` で早めに戻った場合でも `ratatui::restore()` が実行されるようにしましょう。
例えば、ループを別の関数に入れ、その後で `restore` を呼びます。
