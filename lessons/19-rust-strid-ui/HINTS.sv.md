# Ledtrådar

## Bestämma om det ska blinka

```rust
ServerMsg::Status { status } => {
    if let Some(old) = &self.status {
        if status.hp < old.hp {
            self.flash_until = Some(now + FLASH);
        }
    }
    self.status = Some(status);
}
```

## En nästlad layout

```rust
let [map_area, target_area] =
    Layout::vertical([Constraint::Min(3), Constraint::Length(3)]).areas(side_area);
```

## Mätaren

```rust
let ratio = (target.hp.max(0) as f64 / target.max_hp as f64).clamp(0.0, 1.0);
let gauge = Gauge::default()
    .block(Block::default().borders(Borders::ALL).title(format!(" {} ", target.name)))
    .ratio(ratio)
    .label(format!("{}/{}", target.hp.max(0), target.max_hp));
```

Skydda dig mot `max_hp == 0` innan du delar.
