# Lesson 23 – Zones

Phase 4 begins: from a room graph to a tile world à la *Caves of Qud*. Every
room becomes a **zone** – a grid of tiles you walk around in, cell by cell.
Walk out through an opening in the edge and you are in the next room's zone.
The room map from lesson 10 lives on as the *world map*.

```
####.####        # wall    . floor    ~ water
#.......#        @ you     * an item  c a creature
#..~*@..#
#.......#        "n" now takes ONE step north.
#########        "go north" still jumps to the next room.
```

A zone is sent to graphical clients **once**. After that the server only
sends what *changed* – a `delta` frame – because a zone of 31×13 cells with
its creatures, sent ten times a second to every player, would be a waste.

## Goals

- Model a 2D world: a grid (`[][]byte`), positions, neighbours, edges.
- Keep game logic and positions in sync everywhere something moves.
- Compute differences: what has been added, moved and removed since last time.
- Extend a protocol with two new frames without breaking old clients.

## Go concepts

### A grid

```go
type Pos struct{ X, Y int }          // X grows east, Y grows south
type Zone struct {
	W, H  int
	Tiles [][]byte                   // Tiles[y][x] – row first!
}
func (z *Zone) At(p Pos) byte        // outside the zone everything is wall
```

`Pos` is a small struct of two ints – it is **comparable**, so it can be a
map key (`map[Pos]bool`) and compared with `==`. You will use that a lot in
the next lessons.

### Lazy initialization

A room's zone is created the first time someone asks for it:

```go
func (w *World) zoneOf(r *Room) *Zone {
	if r.zone == nil {
		r.zone = ... // ParseZone(r.ZoneRows) or a generated one
	}
	return r.zone
}
```

### Diffing

Every player remembers what their client has been told (`p.seen`, a map from
entity ID to `Entity`). After every command and every tick:

```go
now := s.entities(p.Room)
// in now but not in seen       -> added
// in both, but not equal (!=)  -> moved
// in seen but not in now       -> removed
```

`Entity` is a struct of strings and ints, so `old != e` compares every field.

## The contract

**Tiles** (constants in `zone.go`): `.` floor, `#` wall, `"` grass, `T` tree,
`~` water, `^` rock, `+` door, `<` stairs up, `>` stairs down. `Walkable`: floor,
grass, door, stairs. `tileMeanings` is the legend.

**The world file:**
- a room may have `"zone": ["####.####", ...]` – a hand-made zone (at least
  3×3, equal rows, known tiles – otherwise `MsgErrZoneRows`); every exit needs
  an opening (a walkable edge cell on that side, or stairs for up/down –
  otherwise `MsgErrZoneExit`);
- `"zone_size": [w, h]` – the size of generated zones (default 31×13).

**Rooms without a hand-made zone** get a plain one: walls around a floor, an
opening in the middle of each edge that has an exit, `<` at `(w/2, h/2-1)`
for up and `>` at `(w/2, h/2+1)` for down.

**Positions:** `Player.Pos`, `NPC.Pos`, `Item.Pos`.
- New players, respawned players and restored players: the free cell nearest
  the middle (`freeNear`).
- Through an exit: one step inside the opening you came in by (`arrival`); up
  and down: on the other stairs.
- Items are placed near the middle; a dropped item lies where you stand.

**Commands:**
- `north`, `south`, `east`, `west`, `northeast`, … (`n`, `ne`, …): one step. A
  wall, a tree, water, … gives `MsgBlocked`. Stepping out of the zone through
  an opening in a direction that has an exit takes you to the next room.
- `up`/`down`: only on the stairs (`MsgNoStairs`).
- `go <direction>`: straight to the next room, like before (`MsgGoWhere`, `MsgCantGo`).
- `zone`: the zone again.

**Frames** (see PROTOCOL.md):
- `zone` – `{"room", "width", "height", "rows", "legend", "entities", "self"}`;
  an entity is `{"id", "kind", "name", "glyph", "x", "y"}` with IDs `p:Anna`,
  `n:cat`, `i:lamp`, kinds `player`/`creature`/`item`, glyphs `@`, the NPC's
  `glyph` (default: the first letter of its ID) and the item's `glyph` (default `*`).
- `delta` – `{"room", "added", "moved", "removed"}`; `removed` is a list of IDs.
- Graphical clients get a zone frame with every `look` (and so in every new
  room), and a delta after every command and every tick if something changed.
- Telnet gets the zone drawn as text after every step (`renderZone`, given).

## Task

1. **The fields:** `Room.ZoneRows` (json `zone`) and `Room.zone`;
   `World.ZoneSize` and `worldFile.ZoneSize` (json `zone_size`); `Pos` on
   `Player`, `NPC` and `Item`; `Glyph` (json `glyph,omitempty`) on `NPC` and
   `Item`; `Player.seenRoom` and `Player.seen`; `Frame.Zone` and
   `Frame.Delta`; `case "zone"` in `encodeText`. Now the new files compile.
2. **`zone.go`**: `ParseZone`, `opening`, `arrival`, `plainZone`,
   `openExits`, `zoneOf`, `validateZone` (call it from `Validate`).
3. **`sync.go`**: `entities`, `zoneData`, `sendZone`, `syncAll`, `syncPlayer`.
4. **Moving:** rename `moveCommand` to `travel(p, dir)` and set `p.Pos`; add
   `step`, the stair commands, `go` and `zone`; register the step directions
   and the aliases `ne`, `nw`, `se`, `sw`. Flee uses `travel`.
5. **Everywhere something gets a room, give it a position too:** `addPlayer`,
   `restore`, `die`, `moveNPC`, `respawnCreatures`, `AddNPCs`, `AddItems`,
   `cmdDrop` and `removePlayer` (the helpers `placeItem` and `taken` are handy).
6. **Sync:** `s.syncAll()` after `dispatch`, at the end of `onTick`, after
   logging in and after logging out. `look` sends `sendZone` to graphical clients.

That is a lot of places – the compiler and the tests will find the ones you
forget. Walk around in telnet when you are done: `./mudschool connect`, `n`, `e`, `zone`.

## Think about

- `Tiles[y][x]` – why row first? What would `Tiles[x][y]` mean for `Rows()`?
- What does a delta cost compared with a whole zone? Count the bytes for one step.
- `syncAll` compares *every* entity of every player's room after every command.
  When would that become too slow, and what would you do then?
- The rest of the game still thinks in rooms (combat, `get`, `say`). Is that a
  problem? What would change if `get` required standing on the item?

## Challenges

- Doors: a `+` that is closed (not walkable) until you `open` it.
- Let creatures take a random step inside the zone now and then.
- Show the zone in `look` for telnet players too, next to the description.
