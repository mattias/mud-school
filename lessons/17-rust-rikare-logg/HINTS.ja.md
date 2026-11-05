# ヒント

## テキストのバリアント

```rust
Text {
    #[serde(default)]
    text: String,
    #[serde(default)]
    kind: Kind,
},
```

`ServerMsg::Text { text }` に対する `match` はすべて `kind`（または `..`）を書く必要が
あります。該当箇所はすべてコンパイラが教えてくれます。

## スタイル付きの行を描く

```rust
Line::styled(l.text.as_str(), style_for(l.kind))
```

## 履歴: 上

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

下はその逆です。いちばん新しい項目より先では `history_pos = None` にして、下書きを
戻します。

## 表示位置を動かさない

`push_log` で `self.scroll > 0` なら、新しい行数を `self.scroll` に足します。その後
範囲を制限します: `self.scroll = self.scroll.min(self.log.len())`。
