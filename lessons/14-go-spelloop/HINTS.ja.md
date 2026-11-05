# ヒント

## ループ

```go
ticker := time.NewTicker(s.tick)
defer ticker.Stop()
for {
    select {
    case f := <-s.events:
        f()
    case <-ticker.C:
        s.ticks++
        s.onTick()
    case <-s.stop:
        return
    }
}
```

## call を使った handle

```go
p := newPlayer(conn)
defer p.close()
if !s.call(func() { s.conns[p] = true }) {
    return
}
defer s.call(func() { delete(s.conns, p) })
...
for !p.quit {
    line, ok := p.readLine()
    if !ok {
        return
    }
    if !s.call(func() { s.dispatch(p, line) }) {
        return
    }
}
```

defer は逆順に実行されます。プレイヤーは `p.close()` の *前に* `s.players` と `s.conns` から
削除されるので、閉じた送信箱に誰かが送ることはありません。

## Shutdown

```go
s.lmu.Lock()                 // s.listeners を守る小さなミューテックス
s.closing.Store(true)        // atomic.Bool – Accept が失敗したとき Serve が確認する
for _, l := range s.listeners { l.Close() }
s.lmu.Unlock()
s.call(func() {
    for p := range s.conns {
        p.Send(MsgShutdown)
        p.conn.SetReadDeadline(time.Now())
    }
})
// s.handlers（Serve で接続ごとに Add(1) する sync.WaitGroup）か ctx を待つ
// その後: s.start(); s.stopOnce.Do(func() { close(s.stop) }); <-s.loopDone
```

停止する前に `s.start()` を呼ぶと、`Serve` が一度も呼ばれていなくても `loopDone` が
必ず閉じられます。

リスナーが閉じられる直前に受け付けた接続は、たちの悪い競合を生みます。2 つのルールで防ぎます:

- `Serve` では、`closing` の確認と `s.handlers.Add(1)` を **`lmu` の下で** 行う。
  そうすれば Shutdown が待ち始めた後に `Add` が起きることはない。
- `handle` では、`if !s.closing.Load()` のときだけ（`s.call` の中で）接続を登録する。
  そうしないと別れのあいさつを受け取れず、Shutdown がそれを永遠に待ってしまう。

Shutdown は何度呼んでもかまいません。2 回目以降は nil を返します。

## GameTime

```go
total := startMinute + ticks
day = int(total/(24*60)) + 1
hour = int(total/60) % 24
minute = int(total % 60)
```
