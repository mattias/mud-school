# Lesson 08 – Rooms

A MUD is a world of **rooms** connected by **exits**. Mathematically it is a
*graph*: the rooms are nodes, the exits are directed edges. In this lesson
the players get a place in the world, can `look` around and walk
`north`/`south`/... – and `say` is only heard in the same room.

```
              [Guest room]
                   | up/down
 [Temple] — [Square] — [Smithy]          ← the village of Mudby
                |
           [Town gate]
```

## Goals

- Work with **pointers**: all players in the same room point at the *same* `*Room`.
- Create structs with struct literals: `&Room{ID: "square", ...}`.
- Write a **variadic** function: `NewWorld(start string, rooms ...*Room)`.
- Use **closures** to create commands: `moveCommand("north")`.
- Build text efficiently with `strings.Builder`.

## Go concepts

### Pointers

```go
r := &Room{Name: "Square"} // r is a *Room – a pointer
p.Room = r                 // p and r point at the same room
p.Room == other.Room       // compares ADDRESSES: are they in the same room?
```

### Closures that "remember"

```go
func moveCommand(dir string) CommandFunc {
    return func(s *Server, p *Player, cmd Command) {
        // dir is still here, long after moveCommand has returned
    }
}

for _, dir := range Directions {
    s.Register(dir, moveCommand(dir))   // six commands, one function
}
```

## Task

1. **`world.go`** (new file): write `NewWorld` and `World.Room`.
   **`defaultworld.go`** (new file, finished) contains `DefaultWorld` – the
   village of Mudby. Feel free to add more rooms!
2. **`player.go`**: give `Player` a field `Room *Room`. It is protected by
   `s.mu` (only read and write it while holding the lock).
3. **`server.go`**:
   - give `Server` a field `world *World`,
   - `NewServerWithWorld(w *World) *Server` – and let `NewServer()` call it
     with `DefaultWorld()`,
   - put new players in the start room, and run `look` right after
     `MsgWelcomeName`.
4. **Commands** (put them in a new file `rooms.go` if you like):
   - `look` sends the room's name, description, exits and who else is there:

     ```
     Square                     ← Name
     A square.                  ← Description
     Exits: north, east         ← MsgExits
     Also here: Bertil          ← MsgHere
     ```

     The exits in the order of `Directions`, with the names in
     `directionNames` (in `messages.go`) separated by `MsgListSep`. No exits:
     `MsgNoExits`. The `MsgHere` line (other players, sorted) is only sent if
     someone else is there.
   - `north`, `east`, `south`, `west`, `up`, `down`: if there is no exit,
     send `MsgCantGo`. Otherwise: those in the old room see `MsgLeaves` with
     `directionTravel[dir]` (`Anna leaves north.`), those in the new one see
     `MsgArrives` (`Anna arrives.`), and Anna herself gets the new room's `look`.
   - `say` is now only heard in the same room. The new command `shout` is
     heard everywhere: `MsgYouShout` / `MsgShouts`.

## A convention for locks

As more and more functions need `s.mu` it gets easy to lock twice by mistake
(Go's mutex is **not** reentrant – you get a deadlock). A common trick is to
let functions that *assume* the lock is already held end in `Locked`:

```go
func cmdLook(s *Server, p *Player, cmd Command) {
    s.mu.Lock()
    defer s.mu.Unlock()
    p.Send(s.describeRoomLocked(p))   // does NOT lock by itself
}
```

## Try it yourself

Log in with two clients, walk around, talk, shout.

## Think about

- Why is `Room.Exits` a `map[string]string` (direction → ID) and not
  `map[string]*Room`? (Hint: think about the next lesson, when the world is
  loaded from a file.)
- Exits are one-way. How would you detect an exit `north` that has no
  matching `south` back?

## Challenges

- `go <direction>` and direction names in your own language.
- `exits`, showing where each exit leads: `north – Tavern`.
