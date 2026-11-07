# Lesson 01 – Hello, TCP!

Every MUD since 1978 starts the same way: a server listening on a port and a
player connecting with `telnet`. In this lesson you build the very first
version of your server – an **echo server** that greets you and repeats
everything you type.

## Goals

- Understand how a Go program is put together: `package`, `import`, `func`.
- Use a `struct` with a **method**.
- Handle errors the Go way: `if err != nil { return err }`.
- Listen on TCP, accept a connection and read lines.

## Go in five minutes

```go
package mud                 // every file belongs to a package

import "fmt"                // import from the standard library

type Server struct{}        // a struct (a record with fields – none yet)

func NewServer() *Server {  // a function returning a pointer to a Server
    return &Server{}
}

func (s *Server) Hello() string {  // a METHOD on *Server – (s *Server) is the "receiver"
    return "hello"
}

func divide(a, b int) (int, error) { // functions can return several values
    if b == 0 {
        return 0, fmt.Errorf("cannot divide %d by zero", a)
    }
    return a / b, nil               // nil means "no error"
}
```

Go has no exceptions. Errors are ordinary values of type `error` that you
return and check. It feels verbose at first, but it makes it very clear where
things can go wrong.

`defer` schedules a call to run when the function returns – perfect for
cleaning up:

```go
func (s *Server) handle(conn net.Conn) {
    defer conn.Close()   // runs however we leave the function
    ...
}
```

## Networking in Go

Opening the port is `main.go`'s job. It creates the *listener* and hands it
to your server:

```go
// main.go – already written, you don't need to change it
l, err := net.Listen("tcp", "127.0.0.1:4000")
srv.Serve(l)
```

`Serve` only *uses* the listener it gets:

```go
conn, err := l.Accept()              // wait for the next client – blocks!
fmt.Fprint(conn, "hi\r\n")           // conn is an io.Writer ...
scanner := bufio.NewScanner(conn)    // ... and an io.Reader
for scanner.Scan() {                 // read line by line
    line := scanner.Text()           // the line without its line break
}
```

> **Don't call `net.Listen` inside `Serve`.** The tests give `Serve` a
> listener on a random free port and connect to *that* one. If `Serve` opens
> its own port, the tests never reach your server – and a second listener on
> the same port fails, often with a confusing `nil pointer dereference`.

> **Why `\r\n`?** The telnet protocol (RFC 854) says a line break is CR LF.
> Many clients cope with `\n`, but we do it right from the start.

## Task

Open `workspace/server/mud/server.go` and fill in the two `TODO`s:

1. `Serve` should call `l.Accept()` in an endless `for` loop. If you get an
   error, `Serve` returns it (that happens e.g. when the listener is closed).
   Otherwise call `s.handle(conn)`.
2. `handle` should close the connection with `defer`, send `Welcome` +
   `"\r\n"` and then echo every line it reads.

> **About `messages.go`:** all text the game shows to players lives in
> `workspace/server/mud/messages.go` – in the language you chose for the
> course. `Welcome` is already there. The tests use the same constants, so
> keep the names, but feel free to change the texts. Code and code comments
> are in English, as is customary in the industry; the instructions are in
> your language.

Run `./mudschool check` to test.

## Try it yourself

```
./mudschool run server            # in one terminal
./mudschool connect               # in another (or: telnet / nc 127.0.0.1 4000)
```

Type a few lines. Then connect a *second* time, from a third terminal, at the same time. What happens?
Why? (We fix that in the next lesson.)

## Think about

- What happens if you write `l, err := ...` *inside* the `for` loop? (`:=`
  declares a **new** variable that hides the parameter `l` – this is called
  *shadowing*. `go vet` does not warn about it by default.)
- Why does `Serve` take the listener as a parameter instead of opening the
  port itself? (Hint: think about the tests – and about the next lessons.)

## Challenges (optional)

- Log every new connection with `log.Printf("new connection from %s", conn.RemoteAddr())`.
- Make the server reply `MsgGoodbye` and disconnect if the client types `quit`.
- Read the documentation: `go doc net.Listener` and `go doc bufio.Scanner`.
