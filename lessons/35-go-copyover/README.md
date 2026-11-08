# Lesson 35 – Copyover

Data, scripts and plugins can be reloaded live. But some changes need new Go
code – a new feature in the game loop, a bug fix in combat. That means a new
binary, and a new binary means a restart. Or does it?

**Copyover** is a trick MUDs have used since the 90s. The running server
starts the **new** version of itself and hands over everything that must
survive: the listening socket, every player's open connection, and the
players' state. Then the old process exits. The players see one line:

```
The world shudders for a moment ... everything is as it was, and yet new.
```

– and keep playing, on new code, without ever being disconnected.

```
old process                                   new process (the new binary)
───────────                                   ────────────────────────────
dup the listener and every connection  ──▶    fd 3 = listener, fd 4, 5, ... = players
write the state (rooms, HP, items) as JSON ─▶ MUD_COPYOVER=/tmp/state.json
start the new binary (ExtraFiles)
close its own copies, exit quietly            net.FileListener(fd 3), net.FileConn(fd 4) ...
                                              adopt the players, Serve the listener
```

## Goals

- Understand **file descriptors**: a connection is a number the operating
  system gives a process – and that number can be passed on to a child process.
- Use `exec.Cmd.ExtraFiles`, `(*net.TCPConn).File()`, `net.FileConn` and `net.FileListener`.
- Split a responsibility in two: "log in" and "play" (`handle` → `play`), so
  that a player can start at "play".
- Hand over state between processes – serialized as JSON.

## Go concepts

### File descriptors

`conn.(*net.TCPConn).File()` returns an `*os.File` with a **duplicate** of the
connection's descriptor – a second number for the same socket. A socket is
only closed when *every* descriptor for it is closed. So:

1. the old process duplicates (`File()`),
2. passes the duplicates to the new process (`cmd.ExtraFiles` – they become
   fd 3, 4, 5 … there),
3. closes its own descriptors – the connection stays open, because the new
   process holds copies,
4. the new process turns the numbers back into Go values:
   `net.FileConn(os.NewFile(4, "anna"))`.

### Two kinds of goodbye

`Shutdown` (lesson 14) says goodbye and logs everyone out. `HandOver` must do
neither: no message, no saving, no "Anna left". A flag, `s.handingOver`, tells
the player goroutines to skip their clean-up.

### The real exec

`copyover_main.go` (given, next to `main.go`) does the operating-system part:
`startNewServer` writes the
state to a temporary file and starts `os.Executable()` – the binary on disk,
so a freshly built binary means new code – with `ExtraFiles` and the
environment variable `MUD_COPYOVER`. When `main` starts with that variable
set, it adopts the files (`adoptCopyover`) instead of opening a new listener.
Wiring that into `main.go` is a few lines – see the hints.

### Windows

Windows has no way to hand a socket to a new process like this:
`(*net.TCPConn).File()` returns an error there, and `@copyover` tells you so.
Try copyover in WSL. The tests that need it are skipped on Windows.

## The contract

- `CopyoverPlayer {Name, Room, X, Y, HP, Items, Structured, File}` and
  `CopyoverState {Ticks, Players}` with JSON names `name`, `room`, `x`, `y`,
  `hp`, `items`, `structured`, `file`, `ticks`, `players`.
- `(s *Server) PrepareCopyover() (*CopyoverState, []*os.File, error)` – called
  from outside the game loop; `files[0]` is the listener, each player's `File`
  is the index of their connection's file.
- `(s *Server) HandOver(ctx) error` – stops this server without goodbyes, without
  logging anyone out and without saving; the connections stay open for the new process.
- `(s *Server) Adopt(state, files) (net.Listener, error)` – the new server takes
  the players back: same room (or the start room), position (if still
  walkable), HP and items; each gets `MsgCopyoverDone` and a `look`, and plays
  on without logging in. It returns the listener to `Serve`.
- `(s *Server) Done() <-chan struct{}` – closed when the server has stopped.
- `@copyover` (administrators): with `Options.OnCopyover` set, prepare → start
  the new server → hand over; on an error the administrator gets `MsgCopyoverFailed`.
- `handle` is split: logging in in `handle`, everything after in `play(p)`.

## Task

1. Split `handle` into `handle` and `play`, and let `play`'s clean-up do nothing
   while `s.handingOver` is set.
2. `copyover.go` is partly given (the types, `fileOf`, `Done`, `cmdCopyover`,
   `copyover`): write `PrepareCopyover`, `HandOver`, `Adopt` and `adopt`.
3. `Options.OnCopyover`, the fields (`onCopyover`, and `handingOver` and
   `copyoverRunning` as `atomic.Bool` – the given `cmdCopyover` uses the latter
   so that only one copyover runs at a time) and the command.
4. Wire `copyover_main.go` (given) into `main.go`: `OnCopyover: startNewServer`,
   `adoptCopyover` when `MUD_COPYOVER` is set, and wait for `srv.Done()` too.

Try it for real (Linux, macOS or WSL):

```
cd workspace/server
go build -o mudserver . && ./mudserver -admin Anna
# log in as Anna in another terminal, change a message in messages.go, then:
go build -o mudserver .
# and in the game: @copyover
```

(Use a built binary, not `go run`: `go run` builds a temporary binary and
deletes it – the new process would start the *old* code.)

## Think about

- What happens to a command a player types *during* the copyover?
- What state is *not* handed over (fights, explored zones, creatures' positions, …)?
  Which of it would players notice?
- What if the new binary crashes when it starts? How could the old one notice and carry on?
- Why `os.Executable()` and not `os.Args[0]`?

## Challenges

- Hand over more: fights, the creatures' positions, the clock.
- A health check: the new process says "I'm up" on a pipe before the old one exits;
  otherwise the old one keeps running.
- Hot code reloading the Erlang way is something else entirely – read about it and compare.
