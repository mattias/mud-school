# レッスン 17 – 豊かなログ

レッスン 15 から、サーバーは送るテキストが *どんな種類か* を伝えています:
`{"type":"text","text":"Bertil:「やあ」","kind":"say"}`。今度はクライアントがそれを
活かします。発言は青、エラーは赤、部屋の説明は明るく、人の出入りは灰色。さらに、
どの MUD クライアントにもある 2 つの機能、**コマンド履歴**（上下の矢印キー）と
**スクロール**（PageUp/PageDown）を加えます。

## 目標

- enum とプロトコルを後方互換に拡張する（`#[serde(default)]`、`#[serde(other)]`）。
- enum のバリアントに `#[default]` を付けて `#[derive(Default)]` を使う。
- 小さな状態機械（履歴の位置、スクロール量）を作ってテストする。
- スライスと添字を安全に扱う: `saturating_sub`、`min`。

## Rust の概念

### デフォルト値を持つ enum

```rust
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum Kind {
    #[default]
    Info,          // Kind::default() == Kind::Info
    Room, Say, Move, Error, Combat,
    #[serde(other)]
    Other,         // 新しいサーバーからの "kind":"poetry"
}
```

テキストのバリアントに `#[serde(default)] kind: Kind` を付ければ、`kind` を送らない
レッスン 14 のサーバーでも動き続けます。サーバー側で守ったのと同じ互換性のルールです。

### 安全な添字の計算

```rust
let end = app.log.len().saturating_sub(app.scroll);   // 0 未満にならない
let start = end.saturating_sub(visible);
let lines = &app.log[start..end];
```

`usize` は負になれません。デバッグビルドでは `3 - 5` でパニックします。
`saturating_sub` は代わりに 0 で止まります。

## 契約

- `protocol.rs`: `pub enum Kind { Info, Room, Say, Move, Error, Combat, Other }`
  （デフォルトは `Info`、JSON では小文字、未知のものは `Other`）と
  `ServerMsg::Text { text: String, kind: Kind }`（`kind` のデフォルトは `Info`）。
  普通のテキストの `parse_line` は `Kind::Info` になる。
- `app.rs`:
  - `pub struct LogLine { pub kind: Kind, pub text: String }` と
    `App::log: Vec<LogLine>`。`push_log(&mut self, kind: Kind, text: &str)`。
  - `pub fn style_for(kind: Kind) -> Style` – 種類（`Room`、`Say`、`Move`、`Error`、
    `Combat`）ごとに異なるスタイル。`Other` は `Info` と同じ見た目。ログは各行を
    そのスタイルで描く。
  - 履歴: `pub history: Vec<String>`（古い順、最大 `MAX_HISTORY` 件、同じコマンドが
    続けて 2 回なら 1 回だけ保存）。上で古いコマンド、下で新しいコマンドを表示し、
    いちばん新しいものの先では入力途中だった内容に戻る。
  - スクロール: `pub scroll: usize`（下から何行上か）、`pub const PAGE: usize = 10`。
    PageUp/PageDown で `PAGE` 行ずつ動き、End と Enter で一番下に戻る。上にスクロール
    している間は、新しいメッセージが来ても表示位置は動かない。

## 課題

1. 上の契約どおりに `protocol.rs` と `app.rs` を更新する（`push_log` が種類を取るように
   なったので、`main.rs` も少し変える）。
2. 上にスクロールしていることをログの見出しに表示する。例: `ログ ↑10`。

## 自分で試す

レッスン 16 のサーバーと自分のクライアントを起動します。歩き回り、話し、でたらめを
打ってみましょう。色で何が起きているかわかるはずです。上矢印を何度か押してみて。

## 考えてみよう

- `Kind` は `Copy` なのに `LogLine` はそうでないのはなぜ？
- 将来のサーバーが `"kind":"whisper"` を追加したら、`Kind::Other` はどうなる？
- 履歴とスクロールは端末なしでテストしました。クライアントの他のどこを同じように
  テストできるでしょう？

## チャレンジ

- ログの中の自分の名前を強調する。
- コマンドの Tab 補完（`lo<Tab>` → `look`）。
- 終了時に履歴をファイルへ保存し、起動時に読み込む。
