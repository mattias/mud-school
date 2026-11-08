package mud

import (
	"hash/fnv"
	"math"
	"math/rand/v2"
)

// Procedural generation: zones are built by small algorithms from a seed, so
// the same seed always gives the same world – and nobody has to draw 600 maps.

// GenerateZone builds the zone for a room from its terrain. The same room
// and seed always give exactly the same zone.
func GenerateZone(r *Room, w, h int, seed uint64) *Zone {
	rng := rand.New(rand.NewPCG(seed, hashString(r.ID)))
	var z *Zone
	switch r.Terrain {
	case "cave":
		z = caveZone(w, h, rng, TileWall, TileFloor)
	case "mountain":
		z = caveZone(w, h, rng, TileRock, TileFloor)
	case "forest":
		z = noiseZone(w, h, rng.Uint64(), func(n float64) byte {
			switch {
			case n > 0.5:
				return TileTree
			case n < 0.18:
				return TileWater
			}
			return TileGrass
		})
	case "field":
		z = noiseZone(w, h, rng.Uint64(), func(n float64) byte {
			switch {
			case n > 0.7:
				return TileTree
			case n < 0.15:
				return TileRock
			}
			return TileGrass
		})
	case "water":
		z = noiseZone(w, h, rng.Uint64(), func(n float64) byte {
			if n < 0.55 {
				return TileWater
			}
			return TileGrass
		})
	default: // town, inside, road
		z = NewZone(w, h, TileFloor)
	}
	frame(z, borderTile(r.Terrain))
	openExits(z, r)
	connectZone(z, r, rng, floorTile(r.Terrain), borderTile(r.Terrain))
	return z
}

// borderTile is what surrounds a zone of a terrain.
func borderTile(terrain string) byte {
	switch terrain {
	case "forest", "field":
		return TileTree
	case "mountain":
		return TileRock
	case "water":
		return TileWater
	}
	return TileWall
}

// floorTile is what tunnels are carved with.
func floorTile(terrain string) byte {
	switch terrain {
	case "forest", "field", "water":
		return TileGrass
	}
	return TileFloor
}

// hashString turns a string into a number (FNV-1a), to mix into the seed.
func hashString(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

// frame draws the edge of the zone with tile t.
func frame(z *Zone, t byte) {
	for x := 0; x < z.W; x++ {
		z.Set(Pos{x, 0}, t)
		z.Set(Pos{x, z.H - 1}, t)
	}
	for y := 0; y < z.H; y++ {
		z.Set(Pos{0, y}, t)
		z.Set(Pos{z.W - 1, y}, t)
	}
}

// --- cellular automata -------------------------------------------------

// caveZone: start with random noise (45% wall) and let a cellular automaton
// smooth it into caves.
func caveZone(w, h int, rng *rand.Rand, wall, floor byte) *Zone {
	grid := make([][]bool, h) // true = wall
	for y := range grid {
		grid[y] = make([]bool, w)
		for x := range grid[y] {
			grid[y][x] = x == 0 || y == 0 || x == w-1 || y == h-1 || rng.Float64() < 0.45
		}
	}
	for i := 0; i < 4; i++ {
		grid = caStep(grid)
	}
	z := NewZone(w, h, floor)
	for y := range grid {
		for x := range grid[y] {
			if grid[y][x] {
				z.Tiles[y][x] = wall
			}
		}
	}
	return z
}

// caStep is one step of the "4-5 rule": a cell becomes wall if at least 5 of
// the 9 cells around it (itself included) are walls. Cells outside count as walls.
func caStep(grid [][]bool) [][]bool {
	h, w := len(grid), len(grid[0])
	next := make([][]bool, h)
	for y := range next {
		next[y] = make([]bool, w)
		for x := range next[y] {
			walls := 0
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					ny, nx := y+dy, x+dx
					if ny < 0 || nx < 0 || ny >= h || nx >= w || grid[ny][nx] {
						walls++
					}
				}
			}
			next[y][x] = walls >= 5
		}
	}
	return next
}

// --- noise ---------------------------------------------------------------

// noiseZone fills a zone by looking up value noise for every cell.
func noiseZone(w, h int, seed uint64, pick func(n float64) byte) *Zone {
	z := NewZone(w, h, TileGrass)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			z.Tiles[y][x] = pick(FractalNoise(seed, float64(x)/7, float64(y)/7))
		}
	}
	return z
}

// lattice is a random value in [0,1) for every integer point – the same
// point and seed always give the same value.
func lattice(seed uint64, x, y int) float64 {
	h := seed ^ uint64(int64(x))*0x9E3779B97F4A7C15 ^ uint64(int64(y))*0xC2B2AE3D27D4EB4F
	h ^= h >> 33
	h *= 0xFF51AFD7ED558CCD
	h ^= h >> 33
	h *= 0xC4CEB9FE1A85EC53
	h ^= h >> 33
	return float64(h>>11) / float64(1<<53)
}

// smooth is the "smoothstep" curve 3t²−2t³: no sharp corners between lattice points.
func smooth(t float64) float64 { return t * t * (3 - 2*t) }

// ValueNoise is smooth noise in [0,1): random values on a lattice, blended
// smoothly in between. Points close together get values close together.
func ValueNoise(seed uint64, x, y float64) float64 {
	x0, y0 := math.Floor(x), math.Floor(y)
	tx, ty := smooth(x-x0), smooth(y-y0)
	ix, iy := int(x0), int(y0)
	a := lattice(seed, ix, iy)
	b := lattice(seed, ix+1, iy)
	c := lattice(seed, ix, iy+1)
	d := lattice(seed, ix+1, iy+1)
	top := a + (b-a)*tx
	bottom := c + (d-c)*tx
	return top + (bottom-top)*ty
}

// FractalNoise adds three octaves of value noise – big shapes plus finer
// detail – and keeps the result in [0,1).
func FractalNoise(seed uint64, x, y float64) float64 {
	sum, amp, total := 0.0, 1.0, 0.0
	for o := 0; o < 3; o++ {
		sum += ValueNoise(seed+uint64(o), x, y) * amp
		total += amp
		amp /= 2
		x, y = x*2, y*2
	}
	return sum / total
}

// --- making sure you can walk everywhere ----------------------------------

// connectZone makes sure every opening and staircase can be reached: tunnels are
// carved from the ones that can't to the biggest open area, and cells that
// still can't be reached are filled in.
func connectZone(z *Zone, r *Room, rng *rand.Rand, floor, fill byte) {
	var points []Pos
	for _, dir := range Directions {
		if _, ok := r.Exits[dir]; ok {
			if p, ok := z.opening(dir); ok {
				points = append(points, p)
			}
		}
	}
	main := largestArea(z)
	if len(main) == 0 { // nothing open at all: start with the middle
		z.Set(z.Center(), floor)
		main = reachable(z, z.Center())
	}
	for _, p := range points {
		if !main[p] {
			drunkardWalk(z, p, closest(main, p), rng, floor)
			main = reachable(z, p)
		}
	}
	for y := 1; y < z.H-1; y++ {
		for x := 1; x < z.W-1; x++ {
			if c := (Pos{x, y}); Walkable(z.At(c)) && !main[c] {
				z.Set(c, fill)
			}
		}
	}
}

// neighbours4 are the four cells next to a cell (no diagonals).
var neighbours4 = []Pos{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

// reachable is every walkable cell you can walk to from start (flood fill).
func reachable(z *Zone, start Pos) map[Pos]bool {
	seen := map[Pos]bool{}
	if !Walkable(z.At(start)) {
		return seen
	}
	queue := []Pos{start}
	seen[start] = true
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, d := range neighbours4 {
			n := p.Add(d)
			if !seen[n] && Walkable(z.At(n)) {
				seen[n] = true
				queue = append(queue, n)
			}
		}
	}
	return seen
}

// largestArea is the biggest connected walkable area.
func largestArea(z *Zone) map[Pos]bool {
	var best map[Pos]bool
	done := map[Pos]bool{}
	for y := 0; y < z.H; y++ {
		for x := 0; x < z.W; x++ {
			p := Pos{x, y}
			if done[p] || !Walkable(z.At(p)) {
				continue
			}
			area := reachable(z, p)
			for c := range area {
				done[c] = true
			}
			if len(area) > len(best) {
				best = area
			}
		}
	}
	return best
}

// closest is the cell in area nearest to p (ties broken by position, so the
// result doesn't depend on map order).
func closest(area map[Pos]bool, p Pos) Pos {
	best, found := Pos{}, false
	for c := range area {
		if !found || dist2(c, p) < dist2(best, p) ||
			(dist2(c, p) == dist2(best, p) && (c.Y < best.Y || (c.Y == best.Y && c.X < best.X))) {
			best, found = c, true
		}
	}
	return best
}

// drunkardWalk carves a wobbly tunnel from `from` to `to`: most steps go
// towards the goal, some go anywhere. It never touches the edge of the zone.
func drunkardWalk(z *Zone, from, to Pos, rng *rand.Rand, floor byte) {
	p := from
	if !Walkable(z.At(p)) {
		z.Set(p, floor)
	}
	for steps := 0; p != to && steps < z.W*z.H*8; steps++ {
		var d Pos
		if rng.Float64() < 0.7 {
			d = towards(p, to)
		} else {
			d = neighbours4[rng.IntN(4)]
		}
		n := p.Add(d)
		if n != to && (n.X <= 0 || n.Y <= 0 || n.X >= z.W-1 || n.Y >= z.H-1) {
			continue // keep the frame intact
		}
		p = n
		if !Walkable(z.At(p)) {
			z.Set(p, floor)
		}
	}
	for p != to { // the walk gave up: finish in a straight line
		p = p.Add(towards(p, to))
		if !Walkable(z.At(p)) {
			z.Set(p, floor)
		}
	}
}

// towards is one step (no diagonals) from p in the direction of to.
func towards(p, to Pos) Pos {
	dx, dy := to.X-p.X, to.Y-p.Y
	if abs(dx) >= abs(dy) && dx != 0 {
		return Pos{sign(dx), 0}
	}
	return Pos{0, sign(dy)}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	}
	return 0
}
