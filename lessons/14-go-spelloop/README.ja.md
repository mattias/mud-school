# レッスン 14 – ゲームループ

フェーズ 3 の始まりです。誰も何も入力していなくても世界は動き続けるべきです。時計が
進み、生き物が歩き回り（レッスン 16）、戦闘はラウンドごとに進みます（レッスン 18）。
そのためには **ゲームループ** が必要です。サーバーがデータをどう守るかを考え直す
絶好の機会でもあります。

これまでは、世界に触れるたびにどのゴルーチンも `s.mu` をロックしていました。動きは
しますが、機能が増えるほど難しくなります。どの関数がロックを握っている？ ティックと
コマンドが同時に走ったら？ 多くのゲームサーバー（そして多くの Go プログラム）は別の
モデルを使います:

> **状態はすべて 1 つのゴルーチンが所有する。** 他のゴルーチンはチャネルで *仕事* を
> 送り、答えが必要なら待つ。これが *アクター* パターンです。「メモリを共有して通信する
> のではなく、通信によってメモリを共有せよ」。

```
 connection goroutines              game loop (one goroutine)
 ─────────────────────              ──────────────────────────
 readLine ── s.call(dispatch) ──►   select {
 readLine ── s.call(dispatch) ──►     case f := <-s.events: f()
                                      case <-ticker.C:      s.ticks++; s.onTick()
 Shutdown ── close(s.stop) ─────►     case <-s.stop:        return
                                    }
```

（左は各接続のゴルーチン、右はゲームループのゴルーチンです。）

このレッスンから支援は少なくなります。渡されるのは契約とテストで、設計はあなた次第です。

## 目標

- 状態を所有するゴルーチンを中心にプログラムを組み直す。
- 複数のチャネルに対する `select` と `time.Ticker` を使う。
- `context.Context` で期限を扱い、サーバーをきれいに停止する。
- オプション構造体で設定する。

## Go の概念

### チャネルで関数を送る

```go
events := make(chan func())
go func() {
    for f := range events {
        f()             // このゴルーチンで 1 つずつ実行される
    }
}()

done := make(chan struct{})
events <- func() { counter++; close(done) }
<-done              // 答えを待つ
```

`loop.go` の `do`（送るだけ）と `call`（終わるまで待つ）はまさにこれで、ループが
止まっている場合の抜け道も付いています。

> **ゲームループの中から `s.call` を呼ばないでください。** ループが自分自身を待つ
> ことになり、デッドロックします。コマンドはすでにループの中で動いているので、
> 世界を直接使います。

### context

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
select {
case <-done:          // 時間内に終わった
case <-ctx.Done():    // あきらめた: ctx.Err() は context.DeadlineExceeded
}
```

## 契約

```go
type Options struct {
    World *World        // nil = DefaultWorld()
    Tick  time.Duration // 0 = DefaultTick（100 ms）
}
func NewServerWithOptions(o Options) *Server  // NewServer/NewServerWithWorld はこれを使う
func (s *Server) Ticks() int64                 // 実行されたティック数
func (s *Server) Shutdown(ctx context.Context) error
var ErrServerClosed                            // Shutdown の後に Serve が返すもの
func GameTime(ticks int64) (day, hour, minute int) // 1 ティック = ゲーム内 1 分。1 日目 08:00 から
```

そして `MsgTime` を送るコマンド `time`。

## 課題

1. **`loop.go`**（新規）: `run`、`Ticks`、`Shutdown` を書く。`do` と `call` は用意済み。
2. **`clock.go`**（新規）: `GameTime` を書き、`commands.go` に `time` を登録する。
3. **`server.go`**:
   - `Options`、`DefaultTick`、`ErrServerClosed`、`NewServerWithOptions` と、ループ用の
     フィールド（`loop.go` のコメント参照）、全接続用の `conns map[*Player]bool`、
     `ticks int64` を追加する、
   - **`mu` を削除する**: `players`、`conns`、`ticks`、`p.Room` はゲームループの所有物になる、
   - `handle` と `login` では、世界に触れる処理をすべて `s.call` で実行する: 接続の登録、
     `addPlayer`、すべてのコマンド（`s.call(func() { s.dispatch(p, line) })`）、
     プレイヤーが去るときの後片付け。1 人のプレイヤーのコマンドが順番に実行され、
     `p.quit` を見る前に設定済みになるよう、`do` ではなく `call` を使う、
   - `Serve` はループを開始し、リスナーを覚えておき（Shutdown で閉じるため）、
     Shutdown の後は `ErrServerClosed` を返す。
4. **`rooms.go`/`commands.go`**: ロックを取り除く。`…Locked` の補助関数は
   「ゲームループの中で動く」という意味になったので名前を変える（`look`、
   `describeRoom`、`playersInRoom`、`broadcastRoom` など）。`who` はループ内専用の
   `playerNames` を使い、公開の `PlayerNames` は `s.call` を使う。
5. **`main.go`**（任意だがおすすめ）: `signal.NotifyContext` で Ctrl+C のときに `Shutdown` を
   呼び、全員にあいさつを送る。`-tick` フラグも追加する。

少なくとも一度は `./mudschool check --race` を実行してください。世界に触れているのが
ループだけであることを、レース検出器が証明してくれます。

## 自分で試す

```
./mudschool run server -tick 50ms
```

ログインして `time` を何度か入力します。そのあとサーバーの画面で Ctrl+C を押し、
クライアントに `MsgShutdown` が届くのを見てみましょう。

## 考えてみよう

- `handle` がコマンドに `s.call` ではなく `s.do` を使ったら、何がまずいでしょう？
- `Shutdown` は `readLine` で止まっているゴルーチンを起こすために `SetReadDeadline` を
  使います。ループがプレイヤーの送信箱を閉じるだけではだめなのはなぜ？
- ミューテックスとアクター、それぞれの利点は？ それでもミューテックスを選ぶのはどんなとき？

## チャレンジ

- ゲーム内で 1 時間ごとに鐘を鳴らす: 分が 0 のとき、`onTick` が全員にメッセージを送る。
- 昼と夜: 22:00〜06:00 は `look` に「暗い。」を加える。
