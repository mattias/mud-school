# Ledtrådar

## writeLoop

```go
func (p *Player) writeLoop() {
    defer close(p.done)
    for msg := range p.outbox {
        io.WriteString(p.conn, msg)
    }
}
```

Starta den med `go p.writeLoop()` i `newPlayer`, efter att kanalerna skapats.

## Send som aldrig blockerar

```go
msg := strings.ReplaceAll(text, "\n", "\r\n") + "\r\n"
select {
case p.outbox <- msg:
default:
    p.conn.Close()
}
```

## Säker stängning

Ingen får anropa `Send` efter `close(p.outbox)`. Därför: ta bort spelaren ur
`s.players` (under låset) *innan* du stänger utkorgen, och låt `broadcast`
hålla låset medan den skickar. Då kan ingen hitta spelaren efter borttagningen.

## Ordningen i handle

```go
defer conn.Close()   // körs sist
p := newPlayer(conn)
defer p.close()      // körs näst sist
...
defer func() {       // körs först
    s.removePlayer(p)
    s.broadcast(p, fmt.Sprintf(MsgLeft, p.Name))
}()
```
