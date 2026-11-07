# Lesson 24 – Procedural generation

The world has 21 rooms, and every zone is a rectangle with walls. Drawing 21
zones by hand is boring; drawing 600 is impossible. Games like *Caves of Qud*,
*Brogue* and *Dwarf Fortress* generate their maps with small algorithms – and
from a **seed**, so that the same world can be recreated exactly.

```
###############.###############     TTTTTTTTTTTTTTT.TTTTTTTTTTTTTTT
###############.###############     T"""""""""TT""""""""""""""TTTTT
#########....##..#....#####.###     T"""""""""""T"""""""""""""TTTTT
########......##.......##....##     T"""""""""T"""""""""""""""TTTTT
###..##......................##     .""""""""TTTT"TTT""""""""TTTTTT
##...........................##     T""""""T""TTTTTTTTT"""""TTTTTTT
###.........###>.......###..###     T"""""""""TTTTT""""""""""TTTTTT
        a cave (cellular automaton)          a forest (noise)
```

## Goals

- Implement three classic algorithms: **cellular automata**, **value noise**
  and the **drunkard's walk**.
- Use a seeded random generator so that everything can be recreated – and tested.
- Guarantee a property of generated content (*you can always walk from one
  exit to another*) with a flood fill.

## Go concepts and algorithms

### Seeds

```go
rng := rand.New(rand.NewPCG(seed, hashString(r.ID)))
```

Every room gets its own stream of random numbers, decided by the world's
seed *and* the room's ID (hashed with `hash/fnv`). Change the seed and the
whole world changes; keep it and the world is identical – on every computer.

### Cellular automata (caves)

Fill the zone with noise – 45% wall – and apply a simple rule a few times:
*a cell becomes wall if at least 5 of the 9 cells around it (itself included)
are walls*. Lonely walls disappear, lonely holes close, and what remains
looks like caves.

```
before:     after one step:
#####       #####
#.#.#       ##.##
#...#   ->  ##.##
#.#.#       ##.##
#####       #####
```

### Value noise (forests, fields, lakes)

Random numbers per cell look like static. **Noise** is randomness that varies
*smoothly*: random values on a coarse grid (the *lattice*), blended in
between. Close points get close values, so trees grow in groves and lakes have
shores.

```go
// a, b, c, d: the lattice values around (x, y); tx, ty: smooth(fraction)
top := a + (b-a)*tx
bottom := c + (d-c)*tx
return top + (bottom-top)*ty
```

*Fractal* noise adds several layers (octaves): big shapes plus finer detail.
Then a threshold decides the tile: `n > 0.5` → a tree.

### The drunkard's walk (tunnels)

Generated maps have a nasty habit: unreachable corners, or an exit sealed in
rock. The fix: find the biggest open area (flood fill), and from every exit
that isn't in it, carve a wobbly tunnel – 70% of the steps towards the goal,
30% anywhere. Cells that still can't be reached are filled in, so nobody is
ever placed inside a sealed pocket.

## The contract

- `GenerateZone(r *Room, w, h int, seed uint64) *Zone` – by terrain:
  `cave` (wall/floor) and `mountain` (rock/floor): cellular automata; `forest`,
  `field`, `water`: noise; `town`, `inside`, `road`: a plain floor. Then a frame
  of the terrain's border tile, the openings for the exits (`openExits`), and
  `connectZone`.
- The same room and seed → exactly the same zone; another seed or room → another.
- In every generated zone, **every walkable cell** can be reached from every
  opening, and the edge is closed except at the openings.
- `caStep(grid [][]bool) [][]bool` – the 4-5 rule (cells outside the grid count as walls).
- `ValueNoise(seed, x, y)` – in `[0,1)`, the same for the same input, smooth.
- `reachable(z, start) map[Pos]bool` – flood fill.
- `drunkardWalk(z, from, to, rng, floor)` – connects `from` and `to` and never
  touches the edge.
- The world file gets `"seed": 42`; rooms without a hand-made zone use
  `GenerateZone` with the world's seed and zone size.

## Task

`gen.go` is given except for `caStep`, `ValueNoise`, `FractalNoise`,
`reachable`, `drunkardWalk` and `connectZone`. Then add `Seed` to `World` and
`worldFile` and let `makeZone` call `GenerateZone`.

Walk around: `./mudschool run server`, `./mudschool connect`, `go west` a few
times, `zone`. Try `-world` with a world file of your own with another seed.

## Think about

- Why is the room ID mixed into the seed? What would happen without it?
- The tests generate 160 zones and check that they are connected. Why is
  that a better test than comparing with a picture of the "right" zone?
- What happens to the world if you change `caStep` a tiny bit? To players' saved positions?
- `connectZone` fills unreachable cells. Could that ever remove the only
  path to something? Why not?

## Challenges

- Rooms and corridors (BSP) for `inside` and `town`: split the zone in two,
  again and again, put a room in each part, connect them.
- Real **simplex** noise or Perlin's gradient noise instead of value noise –
  compare how they look.
- Different seeds per area of the world, so that a "mine" can be regenerated
  every day while the town stays the same.
