# レッスン 44 – 負荷テスト

最後のレッスンです。あなたの MUD はインターネットにあり、守られ、翻訳され、ブラウザで
遊べます。でも何人まで耐えられる？ 推測しないで **測りましょう**。このレッスンでは
**非同期 Rust** で負荷テストを書きます。何百ものボットが接続し、ログインし、歩き回り、
サーバーに ping を送ります。そしてサーバーがどれだけ速く答えたかのレポート。

```
$ cargo run --release -p mud-loadtest -- --bots 200 --seconds 10
200 bots against 127.0.0.1:4000 for 10 s ...
bots:          200 (200 connected, 0 failed, 200 logged in)
commands:      2416 in 10.0 s = 242/s
disconnected:  0
echo answers:  79 – p50 1635.9 ms, p95 3746.5 ms, p99 6643.9 ms, max 6643.9 ms
```

この数字は本物です――そして悪い。先を読みましょう。

## 目標

- **tokio** で非同期 Rust を書く: タスク、`async`/`.await`、チャネル。
- `tokio::select!` で最初に来たものを待ち、Go の `select` と比べる。
- **レイテンシー** をきちんと測る: 平均ではなくパーセンタイル。
- 測ったものを使う: `pprof`（レッスン 43）でボトルネックを見つける。

## Rust の概念

### タスク

```rust
tokio::spawn(run_bot(id, plan.clone(), events.clone(), stop.clone()));
```

tokio の *タスク* はゴルーチンのようなものです。安く（数百バイト）、何千も同時に、少数の
スレッドの上でスケジュールされます。`async fn` は await されるか spawn されるまで動きません。
`.await` は、タスクが止まってほかに譲れる場所です。

### `tokio::select!`

```rust
loop {
    tokio::select! {
        line = lines.next_line() => { /* サーバーからの 1 行 */ }
        _ = sleep_until(next), if logged_in => { /* 次のコマンドの時間 */ }
        _ = stop.changed() => { /* テストは終わり */ }
    }
}
```

Go の `select` です。ただしひとひねり: 枝は *future* で、勝たなかったものは捨てられ
（キャンセルされ）ます。`, if logged_in` は枝をオフにします。Go の nil チャネルのように。

### チャネル

ボットは `mpsc` チャネル（送り手は多数、受け手は 1 つ）で集計係に報告し、集計係は `watch`
チャネル（値は 1 つ、変わったことに気づく受け手は多数）でボットに止まれと言います。
Go はどちらにも 1 種類のチャネルを使い、tokio はパターンごとに 1 つ持っています。

### パーセンタイル

平均は待っているプレイヤーを隠します。「p95 = 40 ms」は、応答の 95 % が 40 ms 以内に来て、
5 % は来なかったという意味です。ボットは `echo ping-7` から応答 `ping-7` までの時間を測ります。

## 契約

- `client/loadtest/` はパッケージ `mud-loadtest`（用意済み: `Cargo.toml` と `src/main.rs`）で、
  ワークスペースのメンバーであり、クライアントの dev 依存。
- `percentile(sorted, p)` –「nearest rank」: 順位 ⌈p/100 × n⌉（1 から数え、1..n に収める）の値。
  値がなければ 0。
- `bot_name(id)` –「Bot」と英字だけ（サーバーはそれ以外を受け付けない）、id ごとに違う。
- `run_bot(id, plan, events, stop)` – 接続し（5 秒の期限。失敗は `Event::Failed` とエラー）、
  `Event::Connected`、`HELLO` と名前を送り、その後 `tokio::select!` で: 最初のゾーンで
  `Event::LoggedIn` と `tutorial skip`、差分を適用し、パスワードの問いに答え、echo の応答で
  時間付きの `Event::Reply`、ログイン後は `think`（＋ゆらぎ）ごとにコマンドと `Event::Sent`、
  停止で `quit`。閉じた接続: `Event::Disconnected`。
- `run(plan) -> Report` – ボットを `plan.ramp` に分散して起動し、`plan.duration` が過ぎるまで
  イベントを集め、ボットを止めて待ち、`latencies_ms` をソートする。非同期でないコードには
  `run_blocking`（用意済み）。

## 課題

1. ワークスペースのメンバーに `"loadtest"` を、`[dev-dependencies]` に
   `mud-loadtest = { path = "loadtest" }` を足す。
2. `loadtest/src/lib.rs` は一部用意されています（`Plan`、`Event`、`Report` とその `Display`、
   `Rng`、`next_command`、`run_blocking`）。`percentile`、`bot_name`、`run_bot`、`run` を
   書きましょう。
3. 制限なし・保存なしでサーバーを起動: `./mudschool run server -db "" -max-per-ip 0 -rate 0`、
   そして `client/` で `cargo run --release -p mud-loadtest -- --bots 50`。次は 200。次は 500。

## 数字が語ること

50 体なら応答は 1、2 ミリ秒で来ます。200 体だと何秒もかかります。サーバーが追いつけない
のです。サーバーを `-debug 127.0.0.1:6060` 付きでも起動し、ボットを走らせ、別の端末で:

```
go tool pprof -top http://127.0.0.1:6060/debug/pprof/profile?seconds=10
```

ボットは全員同じ部屋にいます。コマンドのたびに `syncAll` が全プレイヤーの部屋の見え方を
作り直し、部屋にいるプレイヤーの一覧は、プレイヤーごとに *全* プレイヤーの名前をソートして
作られます。コマンドごとに n² log n の仕事です。10 人なら平気でも、200 人には無理です。
それを直すのが、このコースの最後のチャレンジです。

## Go ↔ Rust

| | Go | Rust（tokio） |
|---|---|---|
| 安いスレッド | `go f()` | `tokio::spawn(f())` |
| いくつかのうち最初を待つ | `select { case …: }` | `tokio::select! { … => … }` |
| case をオフにする | nil チャネル | `, if 条件` |
| チャネル | `chan T` | `mpsc`（多 → 1）、`watch`（1 つの値 → 多）、`broadcast`、`oneshot` |
| タイマー | `time.After(d)` | `sleep(d)`、`sleep_until(t)` |

## 考えてみよう

- ボットはパスワードの問いに答えます。それはいつ来る？ `-db ""` なしで負荷テストをした後、
  データベースには何が残る？
- ボットの時計は *送った* ときに始まります。それで見落とすものは？（サーバーが行を読む前に
  起きることを考えましょう。）
- 500 体のボットは少数のスレッド上の 500 のタスクです。`std::thread` で 500 体なら
  スレッドはいくつ必要？ それはなぜ大事？
- ボットでいっぱいの 1 部屋は公平なテスト？ 公平なテストはどんなもの？

## チャレンジ

- 1 部屋 500 体にサーバーを耐えさせる: 部屋ごとのプレイヤー一覧を持ち、何か起きた部屋だけ
  同期する。そしてもう一度測る。
- 代わりに WebSocket でログインするボット（`tokio-tungstenite`）、Caddy 越しに。
- 3 つのパーセンタイルの代わりに、応答時間のヒストグラム。

**これで最後のレッスンです。** Go でマルチプレイヤーのゲームサーバーを、Rust でクライアントを
作り上げました。エコーサーバーから、光とマウスのあるタイルの世界、保存、スクリプトとプラグイン、
誰も切断せずに変えられるサーバー、そして最後に、誰でもブラウザで遊べるインターネット上の MUD まで。
次にどこへ向かうかのアイデアは `docs/*/IDEAS.md` にあります。
