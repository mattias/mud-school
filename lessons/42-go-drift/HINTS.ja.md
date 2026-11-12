# ヒント

## loadConfig – 環境変数からの既定値

`loadConfig` の中の小さな補助関数で短く書けます:

```go
var errs []error
str := func(env, def string) string {
	if v, ok := lookup(env); ok {
		return v
	}
	return def
}
dur := func(env string, def time.Duration) time.Duration {
	v, ok := lookup(env)
	if !ok {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		errs = append(errs, fmt.Errorf("%s=%q: %w", env, v, err))
		return def
	}
	return d
}
fs := flag.NewFlagSet("mudserver", flag.ContinueOnError)
fs.StringVar(&c.Addr, "addr", str("MUD_ADDR", "127.0.0.1:4000"), "... [MUD_ADDR]")
```

グローバルな `flag.Parse` ではなく専用の `flag.FlagSet` を使うのが、テストできる理由です。
呼ぶたびに最初から始まります。最後に `errors.Join(errs...)`。

## main

```go
cfg, err := loadConfig(os.Args[1:], os.LookupEnv)
if err != nil {
	log.Fatal(err)
}
if cfg.HealthCheck {
	if err := healthCheck(cfg.Web); err != nil {
		log.Fatal(err)
	}
	return
}
slog.SetDefault(newLogger(cfg.Log, os.Stderr))
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
```

## HealthHandler

```go
limit := time.After(2 * time.Second)
done := make(chan int, 1)
select {
case s.events <- func() { done <- len(s.players) }:
	select {
	case n := <-done: // 健康
	case <-limit: // the loop took the question but never answered
	}
case <-limit: // the loop is stuck: it doesn't even take the question
case <-s.loopDone:
}
```

`s.call` も `s.do` も使いません。どちらもループが仕事を受け取るまで待ちます。止まって
いれば永遠に。時間制限付きの `select` で自分で `s.events` に送りましょう（`do` と同じく先に
`s.start()` を呼ぶ）。答えのチャネルには 1 つ分の余裕があるので、ループもそこで止まりません。

## TrustProxy

```go
func forwardedFor(r *http.Request) net.IP // 最後の X-Forwarded-For の最後のアドレス
type addrConn struct {
	net.Conn
	remote net.Addr
}
func (c addrConn) RemoteAddr() net.Addr { return c.remote }
```

`net.Conn` を埋め込むと `addrConn` はそのメソッドをすべて持ちます。置き換えるのは 1 つだけです。

## Docker が "unhealthy" と言う

`docker inspect --format '{{json .State.Health}}' <コンテナ>` で `-healthcheck` の答えが見られます。
