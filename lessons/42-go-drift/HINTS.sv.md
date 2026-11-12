# Tips

## loadConfig – standardvärden från miljön

Små hjälpfunktioner inne i `loadConfig` håller den kort:

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

En egen `flag.FlagSet` (inte den globala `flag.Parse`) är det som gör den
testbar: varje anrop börjar från början. Till sist: `errors.Join(errs...)`.

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
done := make(chan int, 1)
if s.do(func() { done <- len(s.players) }) {
	select {
	case n := <-done: // frisk
	case <-time.After(2 * time.Second): // loopen har fastnat
	}
}
```

`s.do`, inte `s.call`: `call` skulle vänta för evigt på en loop som fastnat.
Kanalen har plats för ett värde, så loopen blockeras aldrig av den heller.

## TrustProxy

```go
func forwardedFor(r *http.Request) net.IP // den sista adressen i den sista X-Forwarded-For
type addrConn struct {
	net.Conn
	remote net.Addr
}
func (c addrConn) RemoteAddr() net.Addr { return c.remote }
```

Att bädda in `net.Conn` ger `addrConn` alla dess metoder; du byter bara ut en.

## Docker säger "unhealthy"

`docker inspect --format '{{json .State.Health}}' <container>` visar vad
`-healthcheck` svarade.
