# ヒント

## 点滅するかを決める

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

## 入れ子のレイアウト

```rust
let [map_area, target_area] =
    Layout::vertical([Constraint::Min(3), Constraint::Length(3)]).areas(side_area);
```

## ゲージ

```rust
let ratio = (target.hp.max(0) as f64 / target.max_hp as f64).clamp(0.0, 1.0);
let gauge = Gauge::default()
    .block(Block::default().borders(Borders::ALL).title(format!(" {} ", target.name)))
    .ratio(ratio)
    .label(format!("{}/{}", target.hp.max(0), target.max_hp));
```

割り算の前に `max_hp == 0` の場合に備えましょう。
