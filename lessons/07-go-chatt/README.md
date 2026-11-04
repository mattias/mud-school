# Lesson 07 – Chat

Now it gets social: `say hello` should be heard by everyone else. It sounds
simple – loop over the players and write to their connections – but it hides
a trap that has caught many real MUDs.

## The trap

Imagine Ludvig has a bad connection (or a buggy client that stopped reading).
You type `say hello`. Your goroutine tries to write to Ludvig's socket. The
operating system's buffer for Ludvig is full, so `Write` **blocks** – until
Ludvig reads. While your goroutine waits it may be holding the lock `s.mu`.
Now nobody can log in, nobody can type `who` ... *one* slow player has frozen
the whole server.

The solution is a classic Go pattern:

> **Every connection has exactly one goroutine that writes to it.** Everyone
> else puts messages into a **channel** (the player's outbox) and moves on
> immediately.

```
 Anna's goroutine ──┐
 Bertil's goroutine ┼──► [ Ludvig's outbox (chan string) ] ──► Ludvig's writer goroutine ──► socket
 ...                ┘
```

## Goals

- Create and use **channels**, both unbuffered and buffered.
- Use `select` with `default` to send *without blocking*.
- Close a channel and understand what `range` over a channel does.
- Understand backpressure: what do you do when the receiver can't keep up?

## Go concepts

```go
ch := make(chan string, 256)  // buffered: holds 256 messages before send blocks

ch <- "hi"                    // send (blocks if the buffer is full)
msg := <-ch                   // receive (blocks if the channel is empty)

for msg := range ch { ... }   // receive until the channel is closed
close(ch)                     // only the SENDER closes; sending on a closed channel = panic!

select {                      // pick the first one that is ready ...
case ch <- msg:
    // sent
default:
    // ... or do this at once if nothing is ready
}
```

The channel `done chan struct{}` is an idiom for "a signal without data":
`close(done)` wakes everyone waiting on `<-done`.

## Task

1. **`player.go`**:
   - give `Player` the fields `outbox chan string` (buffered, e.g. 256) and
     `done chan struct{}`,
   - start a `writeLoop` goroutine in `newPlayer` that is the **only** one
     writing to `conn`, and that closes `done` when it finishes,
   - let `Send` put the message in the outbox with `select`/`default`. If the
     outbox is full: disconnect the player (`p.conn.Close()`),
   - write `close()`, which closes the outbox and waits for `done`.
2. **`server.go`**:
   - write `broadcast(except *Player, text string)`, which sends to everyone
     logged in except `except`,
   - tell the others when someone connects (`MsgJoined`: `Bertil has connected.`)
     and leaves (`MsgLeft`: `Bertil has left.`) – whether with `quit` or by
     disconnecting,
   - call `p.close()` before the connection is closed, otherwise `MsgGoodbye`
     may never be sent.
3. **`commands.go`**: add `say`. The speaker sees `MsgYouSay`
   (`You say: hello`), everyone else `MsgSays` (`Anna says: hello`). Without
   text: `MsgSayWhat` (`Say what?`).

The test `TestSlowClientDoesNotBlockOthers` logs in a player who never reads
and lets another one say ~8 MB of text. With the naive solution the test
hangs; with outboxes it passes.

## Try it yourself

Start the server and two clients (telnet or your Rust client from lesson 06)
and talk to yourself. Run `./mudschool check --race` too.

## Think about

- The order of the `defer`s in `handle` matters. What happens if
  `conn.Close()` runs before `p.close()`?
- We disconnect a slow player. What other strategies are there? (Drop
  messages? Merge them? What does a chat app do?)
- An alternative to mutex + channels is to have a *single* goroutine that
  owns the whole game world, to which everyone else sends commands. This is
  sometimes called the "actor" pattern, and we use it in phase 3 for the game loop.

## Challenges

- `tell <name> <text>` – whisper to a single player.
- `emote dances` → `Anna dances.`
