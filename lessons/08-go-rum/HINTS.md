# Ledtrådar

## NewWorld

```go
w := &World{Start: start, Rooms: make(map[string]*Room)}
for _, r := range rooms {
    w.Rooms[r.ID] = r
}
return w
```

## Placera spelaren

I `addPlayer`, som redan håller låset: `p.Room = s.world.Room(s.world.Start)`.

## Spelare i ett rum

```go
func (s *Server) playersInRoomLocked(r *Room, except *Player) []*Player {
    var list []*Player
    for _, other := range s.players {
        if other.Room == r && other != except {
            list = append(list, other)
        }
    }
    // sortera på namn: sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
    return list
}
```

## Gå

```go
to := s.world.Room(p.Room.Exits[dir])  // saknas utgången blir ID "" -> nil
if to == nil {
    p.Send("Du kan inte gå åt det hållet.")
    return
}
```

## Deadlock?

Hänger testerna? Kontrollera att ingen funktion låser `s.mu` medan låset
redan är taget, t.ex. att `cmdSay` inte anropar `s.broadcast` (som låser)
när den redan håller låset.
