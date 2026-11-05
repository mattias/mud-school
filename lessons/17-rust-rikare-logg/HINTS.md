# Hints

## The text variant

```rust
Text {
    #[serde(default)]
    text: String,
    #[serde(default)]
    kind: Kind,
},
```

Every `match` on `ServerMsg::Text { text }` must now mention `kind` (or `..`).
The compiler shows you every place.

## Drawing a styled line

```rust
Line::styled(l.text.as_str(), style_for(l.kind))
```

## History: Up

```rust
let pos = match self.history_pos {
    None if self.history.is_empty() => return,
    None => {
        self.draft = std::mem::take(&mut self.input);
        self.history.len() - 1
    }
    Some(0) => 0,
    Some(p) => p - 1,
};
self.history_pos = Some(pos);
self.input = self.history[pos].clone();
```

Down is the mirror image; past the newest entry, set `history_pos = None` and
put the draft back.

## Keeping the view still

In `push_log`, if `self.scroll > 0`, add the number of new lines to
`self.scroll`. Then clamp: `self.scroll = self.scroll.min(self.log.len())`.
