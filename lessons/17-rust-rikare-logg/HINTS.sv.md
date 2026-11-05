# Ledtrådar

## Text-varianten

```rust
Text {
    #[serde(default)]
    text: String,
    #[serde(default)]
    kind: Kind,
},
```

Varje `match` på `ServerMsg::Text { text }` måste nu nämna `kind` (eller `..`).
Kompilatorn visar dig varje ställe.

## Rita en rad med stil

```rust
Line::styled(l.text.as_str(), style_for(l.kind))
```

## Historik: upp

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

Ner är spegelbilden; förbi den nyaste posten sätter du `history_pos = None` och
lägger tillbaka utkastet.

## Hålla vyn stilla

I `push_log`: om `self.scroll > 0`, lägg till antalet nya rader till
`self.scroll`. Begränsa sedan: `self.scroll = self.scroll.min(self.log.len())`.
