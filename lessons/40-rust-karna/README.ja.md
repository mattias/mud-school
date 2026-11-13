# レッスン 40 – コア

レッスン 41 では、あなたの Rust クライアントをブラウザで動かします。書き直しのように
聞こえますが、クライアントの大部分は、どこで動くかを気にしていません。サーバーからの行の
解釈、ゾーンの更新、森の文字と色を選ぶこと、たいまつの光、クリックしたマスへの経路。
どれも端末もウィンドウもネットワークも時計もいりません。

そこでクライアントを 2 つに分けます。**コア**（独自のライブラリ crate、`mud-core`）は
どこでも同じものをすべて持ちます。端末クライアント、ウィンドウ、そして（次のレッスンの）
ブラウザは、その周りの薄い殻になります:

```
client/                      Cargo の *ワークスペース*
├── Cargo.toml               mud-client: 端末とウィンドウ
├── src/                     app, scene, net, config, messages, gfx …
└── core/                    mud-core: protocol, zone, render, tileset,
    ├── Cargo.toml                     light, path – 入出力なし
    └── src/
```

## 目標

- crate を、独自のライブラリを持つ **ワークスペース** に分ける。
- 純粋なロジックと入出力の境目を引き、それをテストで守る。
- Rust の **孤児ルール**（orphan rule）に出会い、回り道をする。
- 時計を読む代わりに、時刻を外から渡す。

## Rust の概念

### ワークスペース

```toml
[dependencies]
mud-core = { path = "core" }

[workspace]
members = ["core"]
```

メンバーは 1 つの `Cargo.lock` と 1 つの `target/` を共有します。`mud-client` は
ほかの crate と同じように、パスで `mud-core` に依存します。

### 再エクスポートで古いパスを残す

```rust
pub use mud_core::{protocol, render, tileset, zone};
```

をクライアントの `lib.rs` に書けば、`crate::protocol::ServerMsg` や
`mud_client::zone::ZoneView` はそのまま使えます。同じ型が、今はコアに住んでいるだけです。

### 孤児ルール

移したあと、2 つのものがコンパイルできなくなります:

```rust
impl fmt::Display for Status { ... }   // Display: std のもの。Status: コアのもの。
impl From<Rgb> for Color { ... }       // From: std。Rgb: コア。Color: ratatui。
```

Rust では、トレイト *か* 型のどちらかを自分の crate が持っているときだけ、その型に
トレイトを実装できます。そうでないと 2 つの crate が同じものを 2 通りに実装でき、
どちらを使えばいいか誰にもわかりません。クライアントはどちらも持っていないので、
普通の関数にします: `status_line(&Status)` と `color(Rgb)`。（コアなら `Display` を
実装 *できます* が、"online" という言葉はクライアントのもの、コースの言語のものです。）

### 数としての時間

コアは `Instant::now()` を呼んではいけません。ブラウザではパニックします。きらめく光は
時刻をミリ秒として外から受け取ります:

```rust
pub fn zone_light(zone: &ZoneView, terrain: &str, x: usize, y: usize, t_ms: u64) -> Rgb
```

端末クライアントは起動してからの時間を渡し、ブラウザは `requestAnimationFrame` の時計を
渡し、テストは好きな値を渡します。

## 契約

- `client/core/` はパッケージ `mud-core`（用意済み: `Cargo.toml` と `src/lib.rs`）で、
  依存は `serde` と `serde_json` だけ。コードは入出力をせず、時計も読まない
  （`std::net`、`std::fs`、`std::thread`、`Instant::now` などはなし）。
- モジュールは `protocol`、`zone`、`render`、`tileset`（そのまま移す。ただし
  `Display for Status` は除く）、`light`（純粋な部分: `Light`、`TORCH`、`ambient`、`light_at`、
  `lights_for`、`shimmer`、`scale`、そして新たに `shade`、`FOG`、`FULL_LIGHT` と
  新しい `zone_light`）、`path`（`mouse.rs` の `Cell`、`walkable`、`STEPS`、`step_command`、
  `find_path`）。
- `zone_light(zone, terrain, x, y, t_ms)`: 見えないマスなら `FOG`、それ以外は環境光と
  たいまつの光。水は `t_ms` できらめく。
- クライアントは `mud-core` に依存し、`protocol`、`render`、`tileset`、`zone` を
  再エクスポートし、その `light` と `mouse` はコアの部分を再エクスポートする。
  古いパスはそのまま使える。
- `scene::status_line(&Status) -> String`（`messages::ONLINE` 付き）、`scene::cell_light` は
  `zone_light` を使い、`app::color(Rgb) -> Color`。

## 課題

1. `client/Cargo.toml` に `mud-core = { path = "core" }` と `[workspace] members = ["core"]` を足す。
2. `protocol.rs`、`zone.rs`、`render.rs`、`tileset.rs` を `core/src/` に移し（git を使って
   いるなら `git mv`）、`impl Display for Status` を取り除く。
3. `light.rs`（パーティクルと揺れは残る）と `mouse.rs`（メニューとツールチップは残る。
   経路は `core/src/path.rs` へ）を分ける。
4. `zone_light` を書き、コンパイラが見つけるものを直す: コアの中の `crate::` パス、
   `status_line`、`color`。

`./mudschool run client` と `./mudschool run gfx` を実行しましょう。前と同じはずです。

## Go ↔ Rust

| | Go | Rust |
|---|---|---|
| 複数のパッケージ、1 つのビルド | フォルダーにパッケージを置いたモジュール | crate を集めたワークスペース |
| メソッドを足せるのは誰 | 型を定義したパッケージだけ | トレイトか型を持つ crate（孤児ルール） |
| 再エクスポート | –（呼ぶ側が新しいパッケージを import） | `pub use` |

## 考えてみよう

- コアに移せたのに移さなかったモジュールは？ `messages.rs` はなぜ移さない？
- 孤児ルールはクライアントでの `impl From<Rgb> for Color` を禁じます。それはどこに *なら*
  置ける？
- 「時計なし」が WebAssembly だけの話ではないのはなぜ？（テストのことを考えましょう。）

## チャレンジ

- `App` のうち端末に関係ない部分をコアに移す: 端末とブラウザの両方が使える `GameState`。
- `cargo build --target wasm32-unknown-unknown -p mud-core` – ビルドできる？ もうできるはずです。
