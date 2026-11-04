# Hints

## writeLoop

```go
func (p *Player) writeLoop() {
    defer close(p.done)
    for msg := range p.outbox {
        io.WriteString(p.conn, msg)
    }
}
```

Start it with `go p.writeLoop()` in `newPlayer`, after the channels are created.

## A Send that never blocks

```go
msg := strings.ReplaceAll(text, "\n", "\r\n") + "\r\n"
select {
case p.outbox <- msg:
default:
    p.conn.Close()
}
```

## Closing safely

Nobody may call `Send` after `close(p.outbox)`. So: remove the player from
`s.players` (under the lock) *before* closing the outbox, and let `broadcast`
hold the lock while it sends. Then nobody can find the player after the removal.

## The order in handle

```go
defer conn.Close()   // runs last
p := newPlayer(conn)
defer p.close()      // runs second to last
...
defer func() {       // runs first
    s.removePlayer(p)
    s.broadcast(p, fmt.Sprintf(MsgLeft, p.Name))
}()
```
