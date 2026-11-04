# Hints

## The file format

Make a struct of its own for the file:

```go
type worldFile struct {
    Start string  `json:"start"`
    Rooms []*Room `json:"rooms"`
}
```

Decode into it and then create the world with `NewWorld(f.Start, f.Rooms...)`.
The three dots "unpack" the slice into variadic arguments.

## Coordinates

A table makes it easy:

```go
var directionDelta = map[string][3]int{
    "north": {0, -1, 0}, "south": {0, 1, 0}, "east": {1, 0, 0},
    "west": {-1, 0, 0}, "up": {0, 0, 1}, "down": {0, 0, -1},
}
```

And for duplicates: `positions := map[[3]int]string{}`.

## Does the terrain exist?

`slices.Contains(Terrains, r.Terrain)` (the `slices` package, Go 1.21+).

## The embedded file

```go
import (
    "bytes"
    _ "embed"
)

//go:embed world.json
var defaultWorldJSON []byte

func DefaultWorld() *World {
    w, err := LoadWorld(bytes.NewReader(defaultWorldJSON))
    if err != nil {
        panic(err)
    }
    return w
}
```
