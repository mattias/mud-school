# Hints

## An endless loop

Go has only one loop statement: `for`. An endless loop is written `for { ... }`.

```go
for {
    conn, err := l.Accept()
    if err != nil {
        return err
    }
    // ...
}
```

## Use the listener you are given

`Serve` gets a listener `l` that is already listening. Don't call
`net.Listen` in `Serve`, and don't write `l, err := ...` in the loop – that
creates a new `l` that hides the one you were given. Use `=` or a new name
for `conn`, e.g. `conn, err := l.Accept()`.

## Writing to the connection

`net.Conn` satisfies the `io.Writer` interface, so you can use
`fmt.Fprint(conn, "text\r\n")` or `conn.Write([]byte("text\r\n"))`.
Don't forget to import `"fmt"` and `"bufio"` at the top of the file.

## Reading lines

```go
scanner := bufio.NewScanner(conn)
for scanner.Scan() {
    line := scanner.Text()
    fmt.Fprint(conn, line+"\r\n")
}
```

`scanner.Scan()` returns `false` when the client disconnects – then the loop
ends, the function returns and your `defer conn.Close()` runs.
