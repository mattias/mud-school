# Hints

## NewWorld

```go
w := &World{Start: start, Rooms: make(map[string]*Room)}
for _, r := range rooms {
    w.Rooms[r.ID] = r
}
return w
```

## Placing the player

In `addPlayer`, which already holds the lock: `p.Room = s.world.Room(s.world.Start)`.

## Players in a room

```go
func (s *Server) playersInRoomLocked(r *Room, except *Player) []*Player {
    var list []*Player
    for _, other := range s.players {
        if other.Room == r && other != except {
            list = append(list, other)
        }
    }
    // sort by name: sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
    return list
}
```

## Walking

```go
to := s.world.Room(p.Room.Exits[dir])  // a missing exit gives ID "" -> nil
if to == nil {
    p.Send(MsgCantGo)
    return
}
```

## Deadlock?

Do the tests hang? Check that no function locks `s.mu` while the lock is
already held, e.g. that `cmdSay` does not call `s.broadcast` (which locks)
while it already holds the lock.
