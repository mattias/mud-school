# Hints

## loadConfig – defaults from the environment

Small helpers inside `loadConfig` keep it short:

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

A `flag.FlagSet` of its own (not the global `flag.Parse`) is what makes it
testable: every call starts fresh. At the end: `errors.Join(errs...)`.

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
	case n := <-done: // healthy
	case <-limit: // the loop took the question but never answered
	}
case <-limit: // the loop is stuck: it doesn't even take the question
case <-s.loopDone:
}
```

Neither `s.call` nor `s.do`: both wait for as long as it takes the loop to
take the work – forever, if it is stuck. Send to `s.events` yourself, in a
`select` with a time limit (and call `s.start()` first, like `do` does). The
answer channel has room for one, so the loop never blocks on it either.

## TrustProxy

```go
func forwardedFor(r *http.Request) net.IP // the last address in the last X-Forwarded-For
type addrConn struct {
	net.Conn
	remote net.Addr
}
func (c addrConn) RemoteAddr() net.Addr { return c.remote }
```

Embedding `net.Conn` gives `addrConn` all its methods; you only replace one.

## Docker says "unhealthy"

`docker inspect --format '{{json .State.Health}}' <container>` shows what
`-healthcheck` answered.
