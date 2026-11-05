# Lesson 14 – The game loop

Phase 3 begins: the world should live even when nobody types anything. The
clock ticks, creatures wander (lesson 16), fights go round by round (lesson
18). That needs a **game loop** – and it is the perfect moment to rethink how
the server protects its data.

Until now every goroutine has locked `s.mu` whenever it touched the world.
That works, but it gets harder with every feature: which functions hold the
lock? What if a tick and a command run at the same time? Many game servers
(and many Go programs) use another model:

> **One goroutine owns all the state.** Everybody else sends it *work* on a
> channel and, if they need an answer, waits for it. This is the *actor*
> pattern – "don't communicate by sharing memory; share memory by
> communicating".

```
 connection goroutines              game loop (one goroutine)
 ─────────────────────              ──────────────────────────
 readLine ── s.call(dispatch) ──►   select {
 readLine ── s.call(dispatch) ──►     case f := <-s.events: f()
                                      case <-ticker.C:      s.ticks++; s.onTick()
 Shutdown ── close(s.stop) ─────►     case <-s.stop:        return
                                    }
```

From this lesson the scaffolding thins out: you get a contract and tests, and
the design is yours.

## Goals

- Restructure a program around a goroutine that owns the state.
- Use `select` over several channels, and `time.Ticker`.
- Use `context.Context` for a deadline and shut a server down cleanly.
- Configure with an options struct.

## Go concepts

### Functions on a channel

```go
events := make(chan func())
go func() {
    for f := range events {
        f()             // runs in this goroutine, one at a time
    }
}()

done := make(chan struct{})
events <- func() { counter++; close(done) }
<-done              // wait for the answer
```

`do` (fire and forget) and `call` (wait for it) in `loop.go` are exactly that,
plus a way out if the loop has stopped.

> **Never call `s.call` from inside the game loop.** The loop would wait for
> itself – a deadlock. Commands already run in the loop, so they use the
> world directly.

### context

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
select {
case <-done:          // finished in time
case <-ctx.Done():    // gave up: ctx.Err() is context.DeadlineExceeded
}
```

## The contract

```go
type Options struct {
    World *World        // nil = DefaultWorld()
    Tick  time.Duration // 0 = DefaultTick (100 ms)
}
func NewServerWithOptions(o Options) *Server  // NewServer/NewServerWithWorld use it
func (s *Server) Ticks() int64                 // how many ticks have run
func (s *Server) Shutdown(ctx context.Context) error
var ErrServerClosed                            // what Serve returns after Shutdown
func GameTime(ticks int64) (day, hour, minute int) // one tick = one game minute, starting day 1 08:00
```

and the command `time`, which sends `MsgTime`.

## Task

1. **`loop.go`** (new): write `run`, `Ticks` and `Shutdown`. `do` and `call` are given.
2. **`clock.go`** (new): write `GameTime` and register `time` in `commands.go`.
3. **`server.go`**:
   - add `Options`, `DefaultTick`, `ErrServerClosed`, `NewServerWithOptions`
     and the loop fields (see the comment in `loop.go`), plus
     `conns map[*Player]bool` for every connection and `ticks int64`,
   - **remove `mu`**: `players`, `conns`, `ticks` and `p.Room` are now owned
     by the game loop,
   - in `handle` and `login`, run everything that touches the world with
     `s.call`: registering the connection, `addPlayer`, every command
     (`s.call(func() { s.dispatch(p, line) })`), and the clean-up when the
     player leaves. Use `call` (not `do`) so that commands for one player
     run in order and `p.quit` is set before you look at it,
   - `Serve` starts the loop, remembers its listener (Shutdown must close
     it) and returns `ErrServerClosed` after Shutdown.
4. **`rooms.go`/`commands.go`**: remove the locks. The `…Locked` helpers
   are now simply "runs in the game loop" – rename them (`look`,
   `describeRoom`, `playersInRoom`, `broadcastRoom`, …). `who` uses the
   loop-only `playerNames`; the exported `PlayerNames` uses `s.call`.
5. **`main.go`** (optional but nice): `signal.NotifyContext` so that Ctrl+C
   calls `Shutdown` and every player gets a goodbye. Add a `-tick` flag.

Run `./mudschool check --race` at least once – the race detector is your
proof that only the loop touches the world.

## Try it yourself

```
./mudschool run server -tick 50ms
```

Log in, type `time` a few times. Then press Ctrl+C in the server window and
watch the clients get `MsgShutdown`.

## Think about

- What would go wrong if `handle` used `s.do` instead of `s.call` for commands?
- `Shutdown` uses `SetReadDeadline` to wake up goroutines blocked in
  `readLine`. Why can't the loop just close the player's outbox?
- Mutex or actor – what are the advantages of each? When would you choose a
  mutex anyway?

## Challenges

- Ring a bell every game hour: `onTick` sends a message to everyone when the minute is 0.
- Day and night: make `look` add "It is dark." between 22:00 and 06:00.
