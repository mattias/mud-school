package mud

import (
	"fmt"
	"sort"
	"strings"
)

// A zone is the inside of a room: a grid of tiles you walk around in, cell by
// cell, like in Caves of Qud. Walk out through an opening in the edge and
// you are in the next room's zone.

// The tile codes. The server sends codes and a legend with their meanings
// (like the map in lesson 10); the client decides how they look.
const (
	TileFloor      = '.'
	TileWall       = '#'
	TileGrass      = '"'
	TileTree       = 'T'
	TileWater      = '~'
	TileRock       = '^'
	TileDoor       = '+'
	TileStairsUp   = '<'
	TileStairsDown = '>'
)

// tileMeanings is the legend for zone frames.
var tileMeanings = map[byte]string{
	TileFloor: "floor", TileWall: "wall", TileGrass: "grass", TileTree: "tree",
	TileWater: "water", TileRock: "rock", TileDoor: "door",
	TileStairsUp: "stairs_up", TileStairsDown: "stairs_down",
}

// Walkable reports whether you can stand on a tile.
func Walkable(t byte) bool {
	switch t {
	case TileFloor, TileGrass, TileDoor, TileStairsUp, TileStairsDown:
		return true
	}
	return false
}

// Transparent reports whether you can see through a tile (lesson 26).
func Transparent(t byte) bool {
	switch t {
	case TileWall, TileTree, TileRock:
		return false
	}
	return true
}

// Pos is a cell in a zone. X grows to the east, Y to the south.
type Pos struct{ X, Y int }

// Add returns p moved by d.
func (p Pos) Add(d Pos) Pos { return Pos{p.X + d.X, p.Y + d.Y} }

// Zone is a grid of tiles: Tiles[y][x].
type Zone struct {
	W, H  int
	Tiles [][]byte
}

// DefaultZoneW and DefaultZoneH are the size of generated zones when the
// world file doesn't say ("zone_size": [w, h]).
const (
	DefaultZoneW = 31
	DefaultZoneH = 13
)

// NewZone creates a w×h zone filled with tile t.
func NewZone(w, h int, t byte) *Zone {
	z := &Zone{W: w, H: h, Tiles: make([][]byte, h)}
	for y := range z.Tiles {
		z.Tiles[y] = []byte(strings.Repeat(string(t), w))
	}
	return z
}

// ParseZone reads hand-made rows like "#####", "#...#".
func ParseZone(rows []string) (*Zone, error) {
	// TODO(lesson 23): at least 3Ã3, every row as wide as the first, only
	// codes from tileMeanings. Return an error that says what is wrong.
	return nil, fmt.Errorf("TODO")
}

// In reports whether p is inside the zone.
func (z *Zone) In(p Pos) bool { return p.X >= 0 && p.Y >= 0 && p.X < z.W && p.Y < z.H }

// At returns the tile at p. Outside the zone everything is wall.
func (z *Zone) At(p Pos) byte {
	if !z.In(p) {
		return TileWall
	}
	return z.Tiles[p.Y][p.X]
}

// Set changes the tile at p (if it is inside).
func (z *Zone) Set(p Pos, t byte) {
	if z.In(p) {
		z.Tiles[p.Y][p.X] = t
	}
}

// Rows returns the tiles as strings, for a zone frame.
func (z *Zone) Rows() []string {
	rows := make([]string, z.H)
	for y, row := range z.Tiles {
		rows[y] = string(row)
	}
	return rows
}

// Center is the middle cell.
func (z *Zone) Center() Pos { return Pos{z.W / 2, z.H / 2} }

// stepDelta is how a position changes for every walking direction.
var stepDelta = map[string]Pos{
	"north": {0, -1}, "south": {0, 1}, "east": {1, 0}, "west": {-1, 0},
	"northeast": {1, -1}, "northwest": {-1, -1}, "southeast": {1, 1}, "southwest": {-1, 1},
}

// StepDirections are the directions you can step in, in a fixed order.
var StepDirections = []string{"north", "northeast", "east", "southeast", "south", "southwest", "west", "northwest"}

// opposite is the direction you come from.
var opposite = map[string]string{
	"north": "south", "south": "north", "east": "west", "west": "east", "up": "down", "down": "up",
}

// opening returns the cell where the exit in direction dir leaves the zone:
// for north/south/east/west the walkable edge cell closest to the middle of
// that edge, for up/down the stairs. ok is false if there is none.
func (z *Zone) opening(dir string) (Pos, bool) {
	// TODO(lesson 23): see the comment above and the README.
	_ = sort.SliceStable
	return Pos{}, false
}

func dist2(a, b Pos) int { return (a.X-b.X)*(a.X-b.X) + (a.Y-b.Y)*(a.Y-b.Y) }

// arrival is where you appear in this zone when you walk in through the exit
// in direction from (the direction you came FROM, e.g. "south" when you walked
// north): one step inside the opening, or on the stairs.
func (z *Zone) arrival(from string) Pos {
	// TODO(lesson 23): opening(from), then one step inwards if that is walkable.
	return z.Center()
}

// freeNear returns the walkable cell closest to p that isn't in taken.
// If every walkable cell is taken, the closest walkable cell is returned.
func (z *Zone) freeNear(p Pos, taken map[Pos]bool) Pos {
	best, bestTaken := Pos{-1, -1}, Pos{-1, -1}
	for y := 0; y < z.H; y++ {
		for x := 0; x < z.W; x++ {
			c := Pos{x, y}
			if !Walkable(z.At(c)) {
				continue
			}
			if !taken[c] && (best.X < 0 || dist2(c, p) < dist2(best, p)) {
				best = c
			}
			if bestTaken.X < 0 || dist2(c, p) < dist2(bestTaken, p) {
				bestTaken = c
			}
		}
	}
	if best.X >= 0 {
		return best
	}
	if bestTaken.X >= 0 {
		return bestTaken
	}
	return p
}

// plainZone is a zone for a room that has no hand-made zone: walls around a
// floor, with openings for the exits and stairs for up/down.
func plainZone(r *Room, w, h int) *Zone {
	// TODO(lesson 23): NewZone with floor, walls around, then openExits.
	return NewZone(w, h, TileFloor)
}

// openExits makes sure every exit of the room has its opening (or stairs).
func openExits(z *Zone, r *Room) {
	// TODO(lesson 23): the middle of each edge for north/south/east/west,
	// TileStairsUp at (w/2, h/2-1) and TileStairsDown at (w/2, h/2+1).
}

// zoneOf returns the room's zone, creating it the first time.
func (w *World) zoneOf(r *Room) *Zone {
	// TODO(lesson 23): build the zone the first time (ParseZone of ZoneRows, or
	// makeZone) and keep it in r.zone.
	return w.makeZone(r)
}

// zoneSize is the size of generated zones.
func (w *World) zoneSize() (int, int) {
	if w.ZoneSize[0] >= 3 && w.ZoneSize[1] >= 3 {
		return w.ZoneSize[0], w.ZoneSize[1]
	}
	return DefaultZoneW, DefaultZoneH
}

// makeZone creates the zone for a room without a hand-made one.
func (w *World) makeZone(r *Room) *Zone {
	zw, zh := w.zoneSize()
	return plainZone(r, zw, zh)
}

// validateZone checks a room's hand-made zone.
func validateZone(r *Room) []error {
	// TODO(lesson 23): MsgErrZoneRows if ParseZone fails, MsgErrZoneExit for
	// every exit without an opening.
	return nil
}
