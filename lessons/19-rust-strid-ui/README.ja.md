# レッスン 19 – クライアントの戦闘

サーバーはラウンドごとに戦い、毎ラウンド後に status を送ります。そこにはあなたの HP と、
レッスン 18 からは `status.target`（戦っている相手と残りの体力）が入っています。今度は
クライアントがそれを表示します。マップの下に体力バー付きの **ターゲットパネル**、
攻撃を受けたら **赤く点滅** するステータスバーです。

点滅は、クライアントで初めて *時間* に依存するものです。300 ms 赤くなり、元に戻る。
奥深くで `Instant::now()` を呼ぶコードはテストしにくいものです。コツはサーバーと同じ、
**時計を注入する** ことです。

```rust
app.on_server_at(msg, now);   // テストは好きな Instant を渡せる
app.is_flashing(now)
draw_at(frame, &app, now)
// 実際に使うときは:
app.on_server(msg);           // = on_server_at(msg, Instant::now())
```

## 目標

- プロトコルに省略可能な入れ子の構造体を追加する（`#[serde(default)]` 付きの `Option<Target>`）。
- ratatui をさらに使う: 入れ子のレイアウトと `Gauge` ウィジェット。
- `Instant` と `Duration` を扱い、時間を注入してコードをテスト可能にする。
- イベントがなくても画面が更新されるようイベントループを変える（`recv_timeout`）。

## Rust の概念

### 時間

```rust
use std::time::{Duration, Instant};
pub const FLASH: Duration = Duration::from_millis(300);

let until = now + FLASH;           // Instant + Duration = Instant
let flashing = now < until;
self.flash_until.is_some_and(|until| now < until)  // Option<Instant>
```

### イベントなしで目を覚ます

メインループは、キー入力かサーバーのメッセージが来るまで `rx.recv()` で待っていました。
点滅は何も起きなくても *終わる* 必要があるので、最大 50 ms だけ待つようにします:

```rust
match rx.recv_timeout(Duration::from_millis(50)) {
    Ok(ev) => { /* 処理する */ }
    Err(RecvTimeoutError::Timeout) => continue,     // 描き直すだけ
    Err(RecvTimeoutError::Disconnected) => break,
}
```

## 契約

- `protocol.rs`: `pub struct Target { pub name: String, pub hp: i32, pub max_hp: i32 }` と
  `Status::target: Option<Target>`（`#[serde(default)]`）。
- `app.rs`:
  - `pub const FLASH: Duration`（300 ms）、`pub const FLASH_BG: Color`、そして `App` の
    `pub flash_until: Option<Instant>`。
  - `on_server_at(&mut self, msg, now)` – status の `hp` が前回より低ければ `now + FLASH` まで
    点滅させる。`on_server(msg)` は `Instant::now()` でこれを呼ぶ。
  - `is_flashing(&self, now) -> bool`。
  - `pub fn draw_at(frame, app, now)`。`draw(frame, app)` は `Instant::now()` でこれを呼ぶ。
    点滅中はステータスバーの背景を `FLASH_BG` にする。
  - ターゲットがいれば、マップの下にパネルを出す。見出しは相手の名前、中身は `hp/max_hp`
    のラベル付きの `Gauge`（負の hp は 0 と表示）。
- `main.rs`: 上記のとおり `recv_timeout` を使う。

## 課題

`protocol.rs`、`app.rs`、`main.rs` を更新します。その後レッスン 18 のサーバーを起動し、
ネズミと戦ってみましょう。

## 考えてみよう

- `is_flashing` が、タイマーでリセットする単なる `bool` フィールドではないのはなぜ？
- ゲームクライアントで時間に依存するものは他に何がある？（アニメーション、点滅する
  カーソル、「x 秒前」……）それらはどうテストする？
- `Gauge::ratio` は 0〜1 の `f64` を取ります。サーバーが `max_hp` より大きい `hp` を
  送ってきたらどうなる？

## チャレンジ

- ダメージの数字が浮かび上がって消えていく演出（開始時刻付きのエフェクトの一覧を持つ）。
- 体力が半分なら自分の HP のハートを黄色、4 分の 1 未満なら赤にする。
- 攻撃を受けたら数フレームだけマップを 1 列ずらして揺らす。
