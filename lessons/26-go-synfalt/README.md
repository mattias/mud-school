# Lesson 26 – Field of view

So far every player gets the whole zone – including what is behind walls,
round corners and in the next room. In a game that is cheating waiting to
happen: a modified client would show everything. So the server decides what
each player **sees** and sends only that.

```
#####################      o  you
#*******************#      *  visible
#*******************#      .  not visible (the pillar's shadow)
#*******************#
#*******************#
#*********o*#.......#  <- the pillar (#) is seen, what is behind it is not
#*******************#
```

Two kinds of knowledge:

- **visible** – what you see *right now* (it changes with every step);
- **explored** – what you have seen *at some point*. Those tiles stay on the
  map, so you remember the room you came from – "fog of war".

## Goals

- Implement **recursive shadowcasting**, the classic roguelike algorithm.
- Use recursion, slopes as floating-point numbers and symmetry (one octant,
  eight times).
- Separate what the server *knows* from what a client may *know*.

## The algorithm

Split the view into eight octants (triangles). In each, scan row by row,
outwards. Keep the *slopes* `start` and `end` of the part of the row that is
still lit (from 1.0 to 0.0 to begin with). When the scan meets a wall:

1. recurse for the rows behind, but only up to the wall's left edge,
2. continue after the wall with a new `start` – the wall's right edge.

When a row ends blocked, the octant is done: everything behind is in shadow.

```go
for j := row; j <= radius; j++ {
	dy := -j
	for dx := -j; dx <= 0; dx++ {
		left := (float64(dx) - 0.5) / (float64(dy) + 0.5)
		right := (float64(dx) + 0.5) / (float64(dy) - 0.5)
		// skip what is outside start..end, mark the cell visible, handle walls
	}
}
```

The eight octants are the same code with the coordinates mirrored and
swapped: `(dx, dy)` in the octant becomes `(dx*xx + dy*xy, dx*yx + dy*yy)` –
the table `octants` is given.

## The contract

- `ComputeFOV(z *Zone, origin Pos, radius int) map[Pos]bool` – the origin is
  visible; cells within `radius` (Euclidean: `dx²+dy² <= radius²`) that a line
  from the origin reaches; walls, trees and rock are visible but block what is
  behind them (`Transparent`, lesson 23); water doesn't block.
- `PlayerSight = 8`.
- `updateFOV(p)` – stores `p.fov`, adds it to `p.explored[roomID]` and returns
  the cells that were new (in reading order: row by row, left to right).
- **Zone frames:** `rows` contains only explored tiles (the rest is `' '`), the
  legend only the codes that occur, and `visible` is the field of view as rows
  of `'1'` and `'0'`. `entities`: only what is visible – and always yourself.
  Telnet sees the same rows.
- **Delta frames** get `tiles` – `[{x, y, code, meaning}]` for tiles seen for
  the first time – and `visible` (all rows) when the field of view has
  changed. Entities that go out of sight are `removed`, ones that come into
  sight `added`. A delta is sent if *anything* changed, also only `visible`.
- What has been explored is remembered per room as long as you are logged in.

## Task

1. `fov.go`: `castLight` and `updateFOV` (`ComputeFOV`, `octants` and
   `visibleRows` are given).
2. `Player`: `fov`, `explored`, `seenVisible`.
3. `sync.go`: `ZoneData.Visible`, `TileChange`, `DeltaData.Tiles` and
   `DeltaData.Visible`; let `zoneData` show only the explored tiles and
   `visibleEntities`; let `syncPlayer` add new tiles and a changed `visible`.

Your Rust client from lesson 25 ignores the new fields (serde ignores unknown
fields) – it will show newly seen tiles from lesson 28.

## Think about

- Why must FOV be computed on the server and not in the client?
- `ComputeFOV` runs for every player after every command and every tick.
  How expensive is it for 31×13? For 200×200 with 100 players? What could you cache?
- Is shadowcasting *symmetric* – if A sees B, does B see A? Try a few cases.
  Why does it matter in a game with ranged attacks?
- Creatures have no field of view yet. In lesson 27 they will need one.

## Challenges

- Light: dark caves where the field of view is 3 unless you carry a torch.
- Let doors (`+`) block the view when they are closed.
- Save `explored` in the database, so the map is remembered between visits.
