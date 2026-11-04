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
