# Lesson 43 – Protection

Your MUD is on the internet. Sooner or later someone will paste a whole
book into it, script a thousand commands a second, open a hundred
connections, or connect and never say a word. None of that may make the
game slow for everybody else – remember that every command runs in **one**
game loop.

```
> say spam
> say spam
... (500 more)
Slow down – you are sending commands too fast.
...
Too many commands – goodbye.
```

## Goals

- Limit a rate with a **token bucket** – and test it without waiting.
- Put limits where they are cheap: *before* a command reaches the game loop.
- Use deadlines against clients that never finish (`SetReadDeadline`).
- Count what happens with `expvar`, and see where the time goes with `pprof`.

## Go concepts

### A token bucket

```
burst ┃■■■■■┃  every command takes a token;
      ┃■■■  ┃  `rate` tokens flow back in per second;
      ┃     ┃  no token – the command is dropped.
```

```go
func (r *RateLimiter) Allow(now time.Time) bool
```

Short bursts are fine (walking five steps fast), a steady flood is not. The
time is a parameter, so the test can say "half a second later" without
sleeping. The limiter lives in the player's own goroutine (`play`) – a
dropped command never reaches the game loop at all.

### Answer floods quietly

A flood of 500 lines must not get 500 answers – that would double it. The
player gets `MsgSlowDown` once per flood, and is disconnected after
`FloodLimit` dropped commands in a row.

### Limits on reading

```go
p.scanner.Buffer(make([]byte, 0, min(1024, max)), max)   // longer lines: ErrTooLong
conn.SetReadDeadline(time.Now().Add(s.loginTimeout))        // log in in time
```

`bufio.Scanner` refuses a line longer than the buffer's limit; that ends
the connection with `MsgLineTooLong`. (The limit is the larger of `max` and
the buffer's capacity – start small.) The read deadline from lesson 14's
shutdown is reused: logging in must happen in time, and the deadline is
removed once it has – unless a shutdown has set one meanwhile.

### Per address

`p.ip` is the host part of `RemoteAddr` – behind Caddy, thanks to
`TrustProxy` (lesson 42), the player's real address. The game loop counts
the connections in `s.conns` with the same `ip` before it lets a new one in.

### `expvar` and `pprof`

```go
var metricCommands = expvar.NewInt("mud_commands_total")
```

`expvar` publishes counters as JSON on `/debug/vars`; `net/http/pprof` adds
`/debug/pprof/`. Both register on `http.DefaultServeMux`, which the server
serves on its own address, `-debug 127.0.0.1:6060` – never on the public one:

```
go tool pprof http://127.0.0.1:6060/debug/pprof/profile?seconds=10
```

shows which functions the CPU is busy in. Lesson 44 will give it something to do.

## The contract

- `Options.MaxPerIP`, `CommandRate`, `CommandBurst`, `MaxLineLength`
  (0 = `DefaultMaxLine`), `LoginTimeout`. Zero means no limit.
- `NewRateLimiter(rate, burst)` starts full; `Allow(now)` refills by the time
  since the last call, never above `burst`.
- Too many connections from one IP: `MsgTooManyConnections`, closed – and
  there is room again when one closes.
- Commands over the rate are dropped; `MsgSlowDown` once per flood;
  `MsgFlooding` and disconnected after `FloodLimit` in a row.
- A line longer than the limit: `MsgLineTooLong`, closed. Not logged in within
  `LoginTimeout`: `MsgLoginTimeout`, closed. Logged in: no time limit.
- `expvar`: `mud_connections_total`, `mud_connections_refused_total`,
  `mud_commands_total`, `mud_commands_dropped_total`, `mud_players`.
- `main`: `-max-per-ip` (5), `-rate` (10), `-burst` (20), `-login-timeout`
  (1m) and `-debug` (off), with `MUD_MAX_PER_IP`, `MUD_RATE`, `MUD_BURST`,
  `MUD_LOGIN_TIMEOUT` and `MUD_DEBUG`.

## Task

1. `mud/protect.go` is partly given (the constants, the counters, `hostOf`):
   write `Allow`, `connsFrom` and `allowCommand`.
2. `server.go`: the options and fields; in `handle` the IP, the scanner's
   limit, the per-IP check and the login deadline – both inside the `s.call`
   that registers the connection, so a shutdown can't slip in between – and
   its removal after the login; in `play` the limiter. Count players in
   `addPlayer`/`removePlayer`.
   `copyover.go` (lesson 35): `adopt` sets `p.ip` and the scanner's limit, and
   counts the player too.
3. `player.go`: `ip`, `limiter`, `dropped`, and `MsgLineTooLong` in `readLine`.
4. `config.go` and `main.go`: the new flags, and the debug server.

## Think about

- Why is the rate limiter in the player's goroutine, and the per-IP count in
  the game loop?
- A school or a café shares one IP address. What does `-max-per-ip 5` mean for them?
- Bcrypt (lesson 20) is slow on purpose. What could an attacker do with a
  thousand logins at once? What would you limit?
- Which numbers would you want on a dashboard?

## Challenges

- A *global* limit: at most N commands per tick in the game loop, fairly shared.
- Ban an address for ten minutes after a flood.
- Export the counters in Prometheus's text format on `/metrics`.
