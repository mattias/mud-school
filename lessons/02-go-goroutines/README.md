# Lesson 02 – Many players at once

Did you try connecting with two telnet windows in the last lesson? The second
player didn't even get the welcome message until the first one disconnected.
The server does one thing at a time. A MUD without other players is pretty
lonely, so let's fix that.

## Goals

- Understand what a **goroutine** is and start one with `go`.
- Understand the difference between *concurrency* and *parallelism*.
- Use `sync.WaitGroup` to wait for several goroutines.
- Run the tests with the *race detector*: `./mudschool check --race`.

## Goroutines

A goroutine is a lightweight thread scheduled by the Go runtime. They are
cheap – a server can have hundreds of thousands. You start one by putting
`go` in front of a function call:

```go
go s.handle(conn)   // returns immediately; handle runs "on the side"
```

Anonymous functions (closures) work too:

```go
for i := 0; i < 3; i++ {
    go func() {
        fmt.Println("hello from goroutine", i)
    }()
}
```

> Since Go 1.22 every iteration of a `for` loop gets its own copy of `i`. In
> older Go versions it was a classic bug that all goroutines saw the same `i`.

To wait until a bunch of goroutines are done, use a `WaitGroup`:

```go
var wg sync.WaitGroup
for _, job := range jobs {
    wg.Add(1)
    go func() {
        defer wg.Done()
        do(job)
    }()
}
wg.Wait()
```

Look at `mudschool_02_test.go` – the tests use exactly that pattern to let 50
clients send at the same time.

## Task

1. Change `Serve` so that every connection is handled in its own goroutine.
2. Log when a client connects and disconnects with `log.Printf`
   (`conn.RemoteAddr()` gives the client's address). The tests don't check
   the logs, but you will want them.

It is a one-word change plus logging – but the concept is the most important
one in the whole Go part of the course.

## Try it yourself

Start the server and connect with three telnet windows at once. Then run
`./mudschool check --race`. The race detector finds places where two
goroutines touch the same data without synchronization. Your goroutines share
no data yet – but just wait for lesson 05!

## Think about

- What happens if a thousand clients connect? How much memory does a
  goroutine cost? (Answer: a few kilobytes of stack that grows when needed.)
- `handle` blocks on `scanner.Scan()` while waiting for the client. Why is
  that perfectly fine now, but not in lesson 01?

## Challenges

- Count the connected clients and log the number: add a plain `clients int`
  field to `Server`, count up at the start of `handle` and down (with
  `defer`) when the client leaves. Run `./mudschool check --race`: you should
  get `WARNING: DATA RACE`, because many `handle` goroutines change the same
  variable at once. Read about `sync/atomic` (`atomic.Int64`) and fix it.
  (Count in `Serve`'s loop instead and the warning disappears – but why? And
  where would you count *down* then? The race detector only reports races
  that actually happen while the tests run.)
