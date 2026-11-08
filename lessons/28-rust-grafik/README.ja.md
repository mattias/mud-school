# レッスン 28 – グラフィカルクライアント

レッスン 10 でサーバーは、見た目ではなく *意味*（`"forest"`、`"wall"`、`"creature"`）を
送ることを覚えました。端末クライアントは意味を色付きの文字に変えました。ここで 2 つ目の
クライアントを加えます。絵を描く専用の **ウィンドウ** です。そしてクライアントの残り
（プロトコル、ネットワーク、キー操作と履歴を持つ App）は **まったく** 同じままです。

```
サーバー  ──JSON──>  protocol.rs  ──>  App  ──>  build_scene  ──>  Scene
                                                                    │
                                  ┌─────────────── Renderer ────────┴────────┐
                                  │                    │                     │
                            TextRenderer             端末                 GfxRenderer
                            （テスト）             （ratatui）          （ウィンドウ、タイル）
```

ウィンドウには小さなゲームライブラリ **macroquad** と **タイルセット** を使います。16×16
の絵を横に並べた 1 枚の PNG（*アトラス*）です。コースは独自のもの（`client/assets/tiles.png`、
CC0、`tools/tileset/main.go` で作成）を描いています。*Dwarf Fortress* のタイルセットと同じ
考え方です。好きなものに差し替えてかまいません。

## 目標

- 何かを *どう* 描くかを隠す **トレイト** を設計し、複数回実装する。
- 画面に *何が* あるか（`Scene`、ただのデータ）と、*どう* 描くかを分ける。ただのデータは
  テストしやすい。
- **Cargo の feature** と省略可能な依存関係、そして 2 つ目のバイナリを使う。
- アトラスからスプライトを描く: index → PNG 内の長方形。

## Rust の概念

### 複数の実装を持つトレイト

```rust
pub trait Renderer {
    fn cell(&mut self, x: u16, y: u16, cell: &SceneCell);
    fn text(&mut self, x: u16, y: u16, text: &str, fg: Rgb);
}

pub fn draw_scene(scene: &Scene, r: &mut impl Renderer) { ... }
```

`draw_scene` は、グリッドに文字を描いているのか、ウィンドウにテクスチャを描いているのかを
知りません。引数の `impl Renderer` は *静的ディスパッチ* です。コンパイラは renderer の型
ごとに `draw_scene` のコピーを作ります。

### Cargo の feature

グラフィックライブラリは大きく、端末クライアントには不要です:

```toml
macroquad = { version = "0.4", optional = true }

[features]
gfx = ["dep:macroquad"]

[[bin]]
name = "mud-gfx"
path = "gfx/main.rs"
required-features = ["gfx"]
```

プログラムは `src/bin/` ではなく `gfx/main.rs` にあります。Cargo は `src/bin/` のプログラムを
自動で見つけるので、macroquad を加える前にビルドしようとしてしまうからです。`[[bin]]` の
項目で Cargo に場所を教えます。

`cargo test` と `cargo run` は macroquad をまったくビルドしません。`cargo run --release
--features gfx --bin mud-gfx` のときだけビルドされ、コード内の
`#[cfg(feature = "gfx")] pub mod gfx;` もそのときだけ存在します。`[package]` に
`default-run = "mud-client"` を加えて、`cargo run` がどちらのプログラムかわかるようにします。

### ついに戦場の霧

レッスン 26 から delta には `tiles`（新しく見えたタイル）と `visible`（視界）がありましたが、
クライアントは無視してきました。ここで `ZoneView::apply` がタイルをゾーンに入れ、`visible` を
保持します。見たことはあるが今は見えないマスは、ウィンドウでも端末でも暗く描きます
（`shade(色, FOG)`）。

## 契約

- `protocol.rs`: `ZoneData::visible: Vec<String>`、`TileChange {x, y, code, meaning}`、
  `DeltaData::tiles` と `DeltaData::visible`（省略 = 空）。
- `zone.rs`: `ZoneView::visible: Vec<Vec<bool>>`（zone フレームから）、`is_visible(x, y)`
  （サーバーが `visible` を送らなければどこでも true）。`apply` は `tiles` を `rows` に、その
  意味を `legend` に入れ、delta に `visible` があれば置き換える。
- `tileset.rs`: `TILE_PX = 16.0`、`ATLAS_COLUMNS = 16`。`tile_sprite(meaning)`: floor 0、
  wall 1、grass 2、tree 3、water 4、rock 5、door 6、stairs_up 7、stairs_down 8、それ以外は
  `None`。`entity_sprite(kind, is_self)`: 自分 9、プレイヤー 10、生き物 11、アイテム 12、
  それ以外 13。`atlas_rect(index)` → ピクセル単位の `(x, y, w, h)`（1 行に 16 枚）。
- `scene.rs`: `SceneCell {x, y, glyph, fg, bg, tile, entity, light}`、`Scene {width, height,
  cells, status, log, input}`、`LOG_LINES = 5`、`FULL_LIGHT`、`FOG`、`shade(色, 光)`（各成分
  × 光 / 255）、`zone_rows(height)` = `height - 2 - LOG_LINES`、そして:
  - `build_scene(app, width, height)` – `status`（ステータスバーのテキスト、または前後の空白を
    除いた `NOT_LOGGED_IN`）、最後の `LOG_LINES` 行のログ、`input` = `"> "` + 表示される
    入力。ゾーンがあれば、カメラの視野（レッスン 25 のカメラ、`width × zone_rows(height)`）の
    **探索済み** のマスごとに 1 セル。視野座標で、文字と色は `render_zone` から、`tile` と
    `entity` はタイルセットから、`light` は見えれば `FULL_LIGHT`、そうでなければ `FOG`。
  - `draw_scene(scene, renderer)` – 0 行目にステータス、セルは 1 行下（`y + 1`）、ゾーンの
    下にログ、最後の行に入力。
  - `TextRenderer` – 文字のグリッドに描く。`lines()`。
- `gfx.rs` と `gfx/main.rs`（用意済み）: ウィンドウの renderer とプログラム。
- `Cargo.toml`: feature、省略可能な依存関係、2 つ目のバイナリ、`default-run`。
  `assets/tiles.png` は用意済み。
- 端末クライアントも、探索済みだが見えないマスを暗くする。

## 課題

1. プロトコルと `zone.rs`（tiles、visible）。
2. `tileset.rs`（`atlas_rect`）と `scene.rs`。
3. `Cargo.toml`。そして実行: `./mudschool run server` と `./mudschool run gfx --name Anna`。
   最初のビルドには少し時間がかかります。

Linux では X11 と OpenGL の開発パッケージが必要なことがあります（Ubuntu:
`sudo apt install libx11-dev libxi-dev libgl1-mesa-dev`）。Windows と macOS では追加は不要です。

ウィンドウは macroquad 内蔵のフォントを使いますが、これはラテン文字しか持っていません。
ログの日本語（や å、ä、ö）は四角になります。チャレンジを参照してください。

## 考えてみよう

- なぜ `Scene`（ただのデータ）は、直接描くコードよりテストしやすいのでしょう？
- `&mut impl Renderer` と `&mut dyn Renderer` の違いは？ `dyn` が必要になるのはいつ？
- ウィンドウは毎秒 60 回すべてを描き直し、端末は何か起きたときだけ。ゲームではなぜそれで
  よい？
- 端末とウィンドウの *両方* を同時に表示するには何が必要？

## チャレンジ

- 日本語とスウェーデン語の文字を持つフォント: `load_ttf_font` と、`mudclient.toml` の
  `font = "..."` 設定（例: *Noto Sans JP* や *UDEV Gothic*）。
- マウスホイールでズーム（`scale` を変える）。
- 別のタイルセット: Kenney の *1-Bit Pack*（CC0）。それ用の `tile_sprite` を書くか、
  `mudclient.toml` でアトラスファイルを選べるようにする。
- クライアントを **Cargo ワークスペース** にまとめる: `mud-core`（プロトコル、App、シーン）、
  `mud-tui`、`mud-gfx`。
