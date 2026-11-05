# Ledtrådar

## Wanderer

```go
if s.rng.Float64() >= b.Chance {
    return
}
var dirs []string
for _, dir := range Directions {
    if _, ok := n.Location.Exits[dir]; ok {
        dirs = append(dirs, dir)
    }
}
if len(dirs) == 0 {
    return
}
dir := dirs[s.rng.IntN(len(dirs))]
```

## Berätta för ett rum

`broadcastRoom(room, nil, kind, text)` – `except` får vara nil, eftersom en
varelse inte är en spelare.

## AddNPCs och beteenden

```go
for _, spec := range n.Behaviors {
    b, known := newBehavior(spec)
    if !known {
        errs = append(errs, fmt.Errorf(MsgErrBehavior, n.ID, spec.Type))
        continue
    }
    n.AddBehavior(b)
}
```

## onTick

```go
for _, n := range s.world.npcsByID() {
    for _, b := range n.behaviors {
        b.Act(s, n)
    }
}
```
