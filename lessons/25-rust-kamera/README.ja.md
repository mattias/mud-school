# レッスン 25 – カメラ

レッスン 23 からサーバーはゾーンと delta を送っていますが、あなたのクライアントは
それを丁寧に無視してきました（`#[serde(other)] Unknown`、レッスン 11）。ここで描くように
します。ゾーンが画面の大部分を占め、ログは下へ移り、矢印キーで歩けます。

ゾーンは既定で 31×13 マスですが、端末よりずっと大きくなることもあります。そのとき
クライアントはその *窓* を見せます。プレイヤーを追い、ゾーンの外は決して見せない
**カメラ** です。

```
ゾーン（幅 40）  ##########################################
カメラ（10）                [ ....@..... ]       注目点 20 -> 15 から
端の近く                                    [ .......@.# ]  端を越えない
```

## 目標

- 構造体のリストを含む入れ子の JSON と、`self`（Rust のキーワード）という名前の
  フィールドを `#[serde(rename = "self")]` でデシリアライズする。
- 差分更新（delta）で状態を最新に保つ。
- 順序が大事なら `BTreeMap`、そうでなければ `HashMap` を使う。
- 符号なしの数で安全に計算する: `saturating_sub`、`min`。
- 入れ子の ratatui レイアウトで、より複雑な画面を組む。

## Rust の概念

### フィールド名としての `self`

```rust
#[derive(Deserialize)]
pub struct ZoneData {
    // ...
    #[serde(rename = "self", default)]
    pub self_id: String,
}
```

### `From`

`impl From<ZoneData> for ZoneView` は、サーバーが送ったものをクライアントが保持する
形に変換します。行は `Vec<Vec<char>>`（だから `rows[y][x]` は UTF-8 文字列中のバイト位置
ではなく 1 文字）、エンティティは ID をキーにしたマップに。

### `saturating_sub`

`usize` は負になれません。`2 - 5` はデバッグビルドで panic します。
`2usize.saturating_sub(5)` は `0`。左端のカメラにちょうど必要なものです:

```rust
focus.saturating_sub(view / 2).min(size - view)
```

### match ガード

矢印キーは 2 つの役割を持ちます。入力欄が空なら歩き、入力中は履歴をたどる:

```rust
KeyCode::Up | KeyCode::Down | KeyCode::Left | KeyCode::Right
    if self.input.is_empty() && self.history_pos.is_none() && !self.secret =>
{
    return walk_command(key.code).map(String::from);
}
KeyCode::Up if !self.secret => self.history_up(),
```

## 契約

- `protocol.rs`: `Entity {id, kind, name, glyph, x, y}`、`ZoneData {room, width, height,
  rows, legend, entities, self_id}`、`DeltaData {room, added, moved, removed}`（リストの
  省略 = 空）。`ServerMsg::Zone { zone }` と `ServerMsg::Delta { delta }`。
- `zone.rs`（一部用意済み）: `From<ZoneData>` を持つ `ZoneView`、`apply(delta) -> bool`
  （別の部屋の delta は無視して `false`）、`self_pos()`、`code_at`、`meaning_at`、
  `entity_at`。そして `camera(focus, view, size) -> usize`: `focus` を中央に、
  `0..=size-view` に収め、ゾーンが収まるなら `0`。
- `render.rs`: `floor`、`wall`、`grass`、`tree`、`rock`、`door` のタイル。
  `entity_cell(kind, glyph, is_self)`。`render_zone(view, x0, y0, w, h)`。
- `app.rs`: `App::zone: Option<ZoneView>`。zone フレームで設定し、delta で更新する。
  ゾーンがあれば画面は上がゾーン、下がログ。入力欄が空なら矢印キーは `"north"`、
  `"south"`、`"west"`、`"east"` を返す（ログにも履歴にも残さない）。入力中は上下キーで
  これまでどおり履歴をたどる。

## 課題

まずプロトコル、次に `zone.rs`、`render.rs`、そしてアプリ。レッスン 24 のサーバーを動かし、
矢印キーで歩き回ってみましょう。端末を小さくして、カメラの動きを見てみましょう。

## 考えてみよう

- なぜ「違う」部屋の delta が届くことがあるのでしょう？（開口部から出るときの
  フレームの順番を考えてみよう。）
- `entities` は `BTreeMap`、`legend` は `HashMap`。なぜ違う？
- `render_zone` はマスごとに `entity_at` を呼び、全エンティティを探します。マスと
  エンティティがいくつになったら問題になる？ どう速くする？
- ゾーンより先に delta が届いたらどうなる？

## チャレンジ

- 斜め移動: テンキーの 7、9、1、3（または *NetHack* のように `y u b n`）。
- ゾーンが画面より小さいとき、ゾーンの終わりに別の色の枠を描く。
- Shift + 矢印で、何か面白いことが起きるまで歩き続ける（「走る」）。
