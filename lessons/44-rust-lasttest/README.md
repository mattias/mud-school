# Lesson 44 – The load tester

The last lesson. Your MUD is on the internet, protected, translated and
playable in a browser. But how many players can it take? Don't guess –
**measure**. In this lesson you write a load tester in **async Rust**:
hundreds of bots that connect, log in, walk around and ping the server, and
a report of how quickly it answered.

```
$ cargo run --release -p mud-loadtest -- --bots 200 --seconds 10
200 bots against 127.0.0.1:4000 for 10 s ...
bots:          200 (200 connected, 0 failed, 200 logged in)
commands:      2416 in 10.0 s = 242/s
disconnected:  0
echo answers:  79 – p50 1635.9 ms, p95 3746.5 ms, p99 6643.9 ms, max 6643.9 ms
```

Those numbers are real – and bad. Read on.

## Goals

- Write async Rust with **tokio**: tasks, `async`/`.await`, channels.
- Wait for whichever comes first with `tokio::select!` – and compare it with Go's `select`.
- Measure **latency** properly: percentiles, not averages.
- Use what you measure: find the bottleneck with `pprof` (lesson 43).

## Rust concepts

### Tasks

```rust
tokio::spawn(run_bot(id, plan.clone(), events.clone(), stop.clone()));
```

A tokio *task* is like a goroutine: cheap (a few hundred bytes), thousands at
once, scheduled on a few threads. An `async fn` doesn't run until it is
awaited or spawned; `.await` is where a task may pause and let others run.

### `tokio::select!`

```rust
loop {
    tokio::select! {
        line = lines.next_line() => { /* a line from the server */ }
        _ = sleep_until(next), if logged_in => { /* time for a command */ }
        _ = stop.changed() => { /* the test is over */ }
    }
}
```

It is Go's `select` – with a twist: the branches are *futures*, and the ones
that didn't win are dropped (cancelled). `, if logged_in` turns a branch off,
like a nil channel in Go.

### Channels

The bots report to the collector over an `mpsc` channel (many senders, one
receiver); the collector tells them to stop with a `watch` channel (one
value, many receivers that see when it changes). Go has one kind of channel
for both; tokio has one per pattern.

### Percentiles

An average hides the players who wait. "p95 = 40 ms" means 95 % of the
answers came within 40 ms – and 5 % didn't. The bots measure the time from
`echo ping-7` to the answer `ping-7`.

## The contract

- `client/loadtest/` is a package `mud-loadtest` (given: `Cargo.toml` and
  `src/main.rs`), a workspace member and a dev-dependency of the client.
- `percentile(sorted, p)` – "nearest rank": the value at rank ⌈p/100 × n⌉
  (counted from 1, kept within 1..n); 0 for no values.
- `bot_name(id)` – "Bot" and letters only (the server accepts nothing else),
  different for every id.
- `run_bot(id, plan, events, stop)` – connects (5 s timeout; `Event::Failed`
  with the error), `Event::Connected`, sends `HELLO` and its name, and then,
  with `tokio::select!`: on the first zone `Event::LoggedIn` and
  `tutorial skip`; deltas are applied; a password prompt is answered; the
  answer to an echo gives `Event::Reply` with the time; when logged in, a
  command every `think` (+ jitter) with `Event::Sent`; on stop, `quit` (only
  if logged in – before that it would be taken as a name).
  A closed connection: `Event::Disconnected`.
- `run(plan) -> Report` – starts the bots spread over `plan.ramp`, collects
  events until `plan.duration` has passed, stops the bots and waits for
  them (5 s at most in all), and sorts `latencies_ms`. `run_blocking` (given) for code that isn't async.

## Task

1. Add `"loadtest"` to the workspace members and `mud-loadtest = { path = "loadtest" }`
   to `[dev-dependencies]`.
2. `loadtest/src/lib.rs` is partly given (`Plan`, `Event`, `Report` and its
   `Display`, `Rng`, `next_command`, `run_blocking`): write `percentile`,
   `bot_name`, `run_bot` and `run`.
3. Start the server without limits and without saving:
   `./mudschool run server -db "" -max-per-ip 0 -rate 0`, and in `client/`:
   `cargo run --release -p mud-loadtest -- --bots 50`. Then 200. Then 500.

## What the numbers say

With 50 bots the answers come within a millisecond or two. With 200 they take
seconds: the server can't keep up. Start the server with `-debug 127.0.0.1:6060`
too, run the bots, and in another terminal:

```
go tool pprof -top http://127.0.0.1:6060/debug/pprof/profile?seconds=10
```

All the bots stand in the same room. After every command `syncAll` builds
every player's view of the room – and the list of players in the room is
made by sorting *all* players' names, for every player. That is n² log n
work per command: fine for ten players, not for two hundred. Fixing it is
the last challenge of the course.

## Go ↔ Rust

| | Go | Rust (tokio) |
|---|---|---|
| a cheap thread | `go f()` | `tokio::spawn(f())` |
| wait for the first of several | `select { case …: }` | `tokio::select! { … => … }` |
| turn a case off | a nil channel | `, if condition` |
| a channel | `chan T` | `mpsc` (many → one), `watch` (one value → many), `broadcast`, `oneshot` |
| a timer | `time.After(d)` | `sleep(d)`, `sleep_until(t)` |

## Think about

- The bots answer a password prompt. When does one come – and what is left in
  the database after a load test without `-db ""`?
- The bots' clock starts when they *send*. What does that miss? (Think about
  what happens before the server reads the line.)
- 500 bots are 500 tasks on a few threads. How many threads would 500 bots
  with `std::thread` need – and why does it matter?
- Is one room full of bots a fair test? What would a fair one look like?

## Challenges

- Make the server handle 500 bots in one room: keep a list of players per
  room, sync only the rooms where something happened – and measure again.
- Bots that log in over WebSocket instead (`tokio-tungstenite`) – through Caddy.
- A histogram of the answer times instead of three percentiles.

**That was the last lesson.** You have built a multiplayer game server in Go
and a client in Rust – from an echo server to a tile world with light and a
mouse, saving, scripts and plugins, a server that can be changed without
anyone being disconnected, and finally a MUD on the internet that anyone can
play in a browser. `docs/*/IDEAS.md` has ideas for where to go next.
