# Ledtrådar

## roll

```go
if d[1] <= d[0] {
    return d[0]
}
return d[0] + s.rng.IntN(d[1]-d[0]+1)
```

## Rundan

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
    // spelaren slår, kanske seger, annars slår varelsen tillbaka, kanske döden
}
```

## Besegrade varelser

Ta inte bort dem ur `World.NPCs` – markera dem (`dead = true`), ta bort dem ur
rummets `NPCs` och kom ihåg `respawnAt = s.ticks + respawn`.
`respawnCreatures` sätter tillbaka dem när det är dags.

## Att dö

Flytta spelaren till `s.world.Room(s.world.Start)`, sätt `HP = MaxHP`, nollställ
`target`, skicka `MsgYouDie` och anropa `s.look(p)`.
