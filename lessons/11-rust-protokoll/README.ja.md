# レッスン 11 – Rust でプロトコル

サーバーは JSON を話せるようになりました。今度はクライアントがそれを理解する
番です。Go では `Type` フィールドを持つ平らな構造体 `Frame` を送りました。Rust では
もっと強力なものを使います。**各バリアントが自分のデータを持つ enum** です。

```rust
pub enum ServerMsg {
    Hello { text: String },
    Text { text: String },
    Map { map: MapData },
    Status { status: Status },
    Unknown,
}
```

`ServerMsg::Map` は *必ず* マップを持ち、`ServerMsg::Text` は *決して* 持ちません。
不正な状態はそもそも表現できません。Go では `Frame{Type: "text", Map: &m}` が文句なく
コンパイルされるのと比べてみてください。

## 目標

- データ付きの enum と `match` で、すべての場合を扱う。
- Cargo で外部の **クレート** を追加する。
- `tag`、`default`、`other` などの属性を使い、`serde` で JSON をデコードする。
- 標準ライブラリのトレイト `Display` を実装する。
- **前方互換性** を考える: 古いクライアントが新しいサーバー相手に落ちてはいけない。

## Rust の概念

### match は網羅的でなければならない

```rust
match msg {
    ServerMsg::Text { text } => println!("{text}"),
    ServerMsg::Map { map } => draw(&map),
    ServerMsg::Status { status } => println!("[{status}]"),
    ServerMsg::Hello { .. } | ServerMsg::Unknown => {}   // .. = フィールドは気にしない
}
```

バリアントを追加してどこかで扱い忘れると、コンパイラが拒否します。プログラムが
大きくなるにつれて、これは Rust の最大の強みの 1 つになります。

*1 つ* の場合だけに関心があるなら `let ... else` が便利です:

```rust
let ServerMsg::Map { map } = msg else {
    return;   // 関数を抜けなければならない
};
```

### クレートと serde

```
cd workspace/client
cargo add serde --features derive
cargo add serde_json
```

```rust
#[derive(Deserialize)]
#[serde(tag = "type", rename_all = "lowercase")]  // {"type":"text", ...} -> Text
enum ServerMsg {
    Text {
        #[serde(default)]   // フィールドがない？ String::default() = "" を使う
        text: String,
    },
    #[serde(other)]         // 未知の "type" はすべてここへ
    Unknown,
}

let msg: ServerMsg = serde_json::from_str(line)?;
```

**Go ↔ Rust:** Go の構造体タグ `` `json:"max_hp"` `` は、Rust では
`#[serde(rename = "max_hp")]` のような属性に相当します。ここではフィールド名が
すでに同じなので不要です。

### トレイト: Display

```rust
impl fmt::Display for Status {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{} · HP {}/{}", self.name, self.hp, self.max_hp)
    }
}
// これで println!("{status}") と status.to_string() が使える
```

**Go ↔ Rust:** Go の `String() string` メソッド（`fmt.Stringer`）と同じ考えですが、
Rust ではトレイトを *明示的に* 実装します。

## 課題

1. `serde` と `serde_json` を追加する（上記参照）。
2. **`src/protocol.rs`**（新しいファイル）: バリアント、構造体、`MapData::meaning`、
   `Status` の `Display`（「オンライン」の語には `messages::ONLINE` を使う）、
   `parse_line` を埋める:
   - `{` で始まり正しい JSON である行は、適切なバリアントになる、
   - それ以外（普通のテキスト、壊れた JSON）は行全体を持つ `ServerMsg::Text` になる。
3. `lib.rs` に `pub mod protocol;` を追加する。
4. **`src/main.rs`**: 接続直後に `HELLO` を送り、メッセージをきれいに表示する:
   テキストはそのまま、マップは行ごと、ステータスは `[...]` で囲む。

## 自分で試す

```
./mudschool run server
./mudschool run client
```

世界を歩き回ってみましょう。一歩ごとにマップが描かれます！

## 考えてみよう

- 未知のフレーム種別をエラーにするより `Unknown` にするほうがよいのはなぜ？
- `parse_line` は `Result<ServerMsg, _>` ではなく `ServerMsg` を返します。それで
  何を得て、何を失うでしょう？

## チャレンジ

- 独自のエラー型 `enum ProtocolError { NotJson, BadFrame(serde_json::Error) }` と
  `fn parse_frame(line) -> Result<ServerMsg, ProtocolError>` を書く。`Display` と
  `std::error::Error` を実装する。
- `Serialize` を追加し、エンコードしてデコードする（往復の）テストを書く。
