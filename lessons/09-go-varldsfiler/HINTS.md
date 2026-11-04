# Ledtrådar

## Filformatet

Gör en egen struct för filen:

```go
type worldFile struct {
    Start string  `json:"start"`
    Rooms []*Room `json:"rooms"`
}
```

Avkoda till den och skapa sedan världen med `NewWorld(f.Start, f.Rooms...)`.
De tre punkterna "packar upp" slicen till variadiska argument.

## Koordinater

En tabell gör det enkelt:

```go
var directionDelta = map[string][3]int{
    "north": {0, -1, 0}, "south": {0, 1, 0}, "east": {1, 0, 0},
    "west": {-1, 0, 0}, "up": {0, 0, 1}, "down": {0, 0, -1},
}
```

Och för dubbletter: `positions := map[[3]int]string{}`.

## Finns terrängen?

`slices.Contains(Terrains, r.Terrain)` (paketet `slices`, Go 1.21+).

## Inbakad fil

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
