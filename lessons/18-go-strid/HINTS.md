# Hints

## roll

```go
if d[1] <= d[0] {
    return d[0]
}
return d[0] + s.rng.IntN(d[1]-d[0]+1)
```

## The round

```go
for _, name := range s.playerNames() {
    p := s.players[name]
    n := p.target
    if n == nil {
        continue
    }
    if n.dead || n.Location != p.Room {
        p.target = nil
        continue
    }
    // player hits, maybe defeat, otherwise the creature hits back, maybe die
}
```

## Defeated creatures

Don't delete them from `World.NPCs` – mark them (`dead = true`), remove them
from their room's `NPCs` and remember `respawnAt = s.ticks + respawn`.
`respawnCreatures` puts them back when the time has come.

## Dying

Move the player to `s.world.Room(s.world.Start)`, set `HP = MaxHP`, clear
`target`, send `MsgYouDie` and call `s.look(p)`.
