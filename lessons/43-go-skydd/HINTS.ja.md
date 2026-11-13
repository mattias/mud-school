# ヒント

## Allow

```go
if !r.last.IsZero() {
	r.tokens += now.Sub(r.last).Seconds() * r.rate
	if r.tokens > r.burst {
		r.tokens = r.burst
	}
}
r.last = now
if r.tokens < 1 {
	return false
}
r.tokens--
return true
```

## allowCommand

```go
if p.limiter == nil || p.limiter.Allow(time.Now()) {
	p.dropped = 0
	metricCommands.Add(1)
	return true, false
}
```

……そうでなければ数えて `p.dropped++`、`p.dropped == 1` のとき `MsgSlowDown`、
`FloodLimit` で `MsgFlooding` と `quit = true`。

## play

```go
if ok, quit := s.allowCommand(p); quit {
	return
} else if !ok {
	continue // 捨てた
}
```

## handle – 接続を断る

接続を登録する `s.call` の中で決め（`s.conns` を読めるのはゲームループだけ）、その後で
`MsgTooManyConnections` を送ります。defer した `p.close()` が、接続が閉じる前に確実に送ります。

## ログインの期限

`p` を `s.conns` に加えるのと同じ `s.call` の中、その直前で設定します。後でループの外で
設定すると、その間の `Shutdown` が先に自分の期限（今）を設定し、それをあなたの 1 分が
上書きしてしまいます。ログインに失敗したら:

```go
if !ok {
	if errors.Is(p.scanner.Err(), os.ErrDeadlineExceeded) && !s.closing.Load() {
		p.SendKind(KindError, MsgLoginTimeout)
	}
	return
}
```

そしてログインに成功したら、ゲームループの中で:
`if !s.closing.Load() { conn.SetReadDeadline(time.Time{}) }`。

## デバッグ用のサーバー

```go
import _ "net/http/pprof"
...
go func() { log.Print(http.ListenAndServe(cfg.Debug, nil)) }() // nil = DefaultServeMux
```
