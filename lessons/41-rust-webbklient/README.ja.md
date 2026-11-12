# レッスン 41 – ブラウザのクライアント

レッスン 39 からブラウザで遊べるようになりました。テキストだけで。今度はブラウザが
マップを手に入れます。レッスン 40 のコアを **WebAssembly** にコンパイルし、ゾーンを
`<canvas>` に描きます。たいまつの光と霧もあり、クリックで歩けます。ウィンドウと同じ A* です。

ページ（`server/web/mud.js`、レッスン 39 から用意済み）は、もう話し方を知っています。
`mud_web.wasm` があれば読み込みます。

```
WebSocket ──行──▶ mud.js ──on_line──▶ mud_web.wasm  （ゾーン、差分、ステータス）
                  mud.js ◀──frame──── mud_web.wasm  （cols × rows マス、毎秒 60 回）
クリック  ──col,row─────────▶ mud_web.wasm ──mud.send("east")──▶ mud.js ──▶ WebSocket
```

## 目標

- Rust を WebAssembly にコンパイルする: `--target wasm32-unknown-unknown` と `cdylib`。
- 2 つの言語の間の小さな **ABI**（数と共有メモリ）を設計し、それがレッスン 33 の
  逆向きだと気づく。
- ロジックはテストしやすく: `WebClient` は普通の Rust。WebAssembly を知っているのは薄い層だけ。
- 別の言語のためにデータを詰める: 決まった並びのバイト。

## Rust の概念

### `cdylib`

```toml
[lib]
crate-type = ["cdylib", "rlib"]
```

`cdylib` は WebAssembly に必要なもの、C 風のインターフェースを持つライブラリです。
`rlib` は普通の Rust ライブラリで、`WebClient` を直接使うテストのためです。

### ABI、バージョン 1

レッスン 33 では Go サーバーがホストで、Rust のプラグインがゲストでした。今度は
ブラウザがホストです:

| エクスポート | |
|---|---|
| `abi_version() -> i32` | 1 |
| `alloc(len) -> ptr`、`dealloc(ptr, len)` | ページが行を書き込むメモリ |
| `on_line(ptr, len)` | サーバーからの 1 行 |
| `frame(cols, rows, t_ms) -> ptr` | 画面: 12 バイトのマスが cols × rows |
| `click(col, row)`、`tick(t_ms)` | クリック、時間の経過 |

そして `mud.send(ptr, len)`（サーバーへのコマンド）をインポートします。マスは 12 バイト:
文字（`u32`、リトルエンディアン）、前景の r g b、背景の r g b、背景があれば 1、予備の 1 バイト。

### `#[cfg(target_arch = "wasm32")]` と `#[no_mangle]`

エクスポートする関数は、WebAssembly 向けにコンパイルするときだけ存在します。
`thread_local!` に `WebClient` を 1 つ持ち（ブラウザのスレッドは 1 つ）、そのメソッドを呼びます:

```rust
#[no_mangle]
pub extern "C" fn frame(cols: usize, rows: usize, t_ms: f64) -> *const u8 {
    CLIENT.with_borrow_mut(|c| c.frame(cols, rows, t_ms as u64).as_ptr())
}
```

`#[no_mangle]` は名前 `frame` を残すので、ページが見つけられます。バイトは呼び出しの後も
`WebClient` に残るので、ページが読むときにもポインターは有効です。

## 契約

- `client/web/` はパッケージ `mud-web`（用意済み: `Cargo.toml`）で、ワークスペースの
  メンバーであり、クライアントの dev 依存（テストが使えるように）。
- `WebClient::new()`、`on_line(&str)`（ゾーン・差分・ステータスを保つ。別の部屋のゾーンは
  歩きを止める。それ以外は無視）、`frame(cols, rows, t_ms) -> &[u8]`、`click(col, row)`、
  `tick(t_ms)`、`take_outbox() -> Vec<String>`。`ABI_VERSION = 1`、`CELL_BYTES = 12`、`WALK_MS`。
- `frame`: `cols × rows × CELL_BYTES` バイト、ゾーンがなければ全部 0。ゾーンがあれば、
  レッスン 25 のカメラでプレイヤーを中心に。フレームより小さいゾーンは中央に置く。
  各マスは `render_zone` から、色は `zone_light`（地形は最後のステータスから）で `shade` する。
- `click(col, row)`: その下のゾーンのマス（最後のフレームでの描画どおり）、`find_path` の経路、
  コマンドとしての歩み。`tick` は `WALK_MS` ごとに多くても 1 つ送る。
- `wasm32-unknown-unknown` 向けにビルドした `mud_web.wasm` は `memory`、`abi_version`、
  `alloc`、`dealloc`、`on_line`、`frame`、`click`、`tick` をエクスポートし、インポートは
  `mud.send` だけ。

## 課題

1. `rustup target add wasm32-unknown-unknown`（レッスン 33 でまだなら）。
2. `client/Cargo.toml` で、ワークスペースのメンバーに `"web"` を、`[dev-dependencies]` に
   `mud-web = { path = "web" }` を足す。
3. `web/src/lib.rs` は一部用意されています（定数、構造体、`take_outbox`、ABI の層）。
   `on_line`、`frame`、`click`、`tick` を書きましょう。
4. `./mudschool run web` は `mud_web.wasm` をビルドして `server/web/` にコピーし、サーバーを
   起動します。http://127.0.0.1:8080 を開きましょう。

## Go ↔ Rust

| | レッスン 33 | レッスン 41 |
|---|---|---|
| ホスト | Go サーバー（wazero） | ブラウザ（mud.js） |
| ゲスト | Rust のプラグイン | Rust のクライアント |
| ゲストへ | `"player\tverb\targs"` | サーバーからの 1 行 |
| ゲストから | `mud.send` – プレイヤーへのテキスト | `mud.send` – コマンド、`frame` – やがてピクセルになるもの |

## 考えてみよう

- なぜ `frame` は新しい `Vec` ではなく `WebClient` の中を指すポインターを返すのでしょう？
  関数の終わりで捨てられる `Vec` だと何がまずい？
- `CELL_BYTES` は 11 ではなく 12 です。JavaScript にとって 4 の倍数がうれしいのはなぜ？
- ページは毎秒 60 回描きますが、ゾーンが変わるのは毎秒数回です。その仕事を省くには？

## チャレンジ

- 文字の代わりにタイルを描く: `tiles.png`（レッスン 28）を配信し、`frame` にスプライト番号も入れる。
- ホバー: `describe_cell` のツールチップ（レッスン 30 – 先にコアへ移す）。
- ブラウザでのパーティクル（レッスン 29）。
