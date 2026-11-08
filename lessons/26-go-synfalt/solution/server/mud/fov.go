package mud

// Field of view: what a player can see from where they stand. It is computed
// on the server – if the client got the whole zone, a modified client could
// see through walls.

// PlayerSight is how far players see (in cells).
const PlayerSight = 8

// octants turn the one-octant algorithm below into all eight directions:
// (x, y) in the octant becomes (x*xx + y*xy, x*yx + y*yy) in the zone.
var octants = [8][4]int{
	{1, 0, 0, 1}, {0, 1, 1, 0}, {0, -1, 1, 0}, {-1, 0, 0, 1},
	{-1, 0, 0, -1}, {0, -1, -1, 0}, {0, 1, -1, 0}, {1, 0, 0, -1},
}

// ComputeFOV returns every cell visible from origin within radius, using
// recursive shadowcasting. Walls you can see are visible; what is behind
// them is not.
func ComputeFOV(z *Zone, origin Pos, radius int) map[Pos]bool {
	vis := map[Pos]bool{origin: true}
	for _, o := range octants {
		castLight(z, origin, radius, 1, 1.0, 0.0, o, vis)
	}
	return vis
}

// castLight scans one octant row by row, outwards from the origin. start and
// end are the slopes of the part of the row that is still lit. When it meets
// a wall it calls itself for the part before the wall, and continues after it.
func castLight(z *Zone, origin Pos, radius, row int, start, end float64, o [4]int, vis map[Pos]bool) {
	if start < end {
		return
	}
	newStart := 0.0
	for j := row; j <= radius; j++ {
		dy := -j
		blocked := false
		for dx := -j; dx <= 0; dx++ {
			p := Pos{origin.X + dx*o[0] + dy*o[1], origin.Y + dx*o[2] + dy*o[3]}
			left := (float64(dx) - 0.5) / (float64(dy) + 0.5)
			right := (float64(dx) + 0.5) / (float64(dy) - 0.5)
			if start < right {
				continue
			}
			if end > left {
				break
			}
			if dx*dx+dy*dy <= radius*radius && z.In(p) {
				vis[p] = true
			}
			opaque := !Transparent(z.At(p))
			if blocked {
				if opaque {
					newStart = right
					continue
				}
				blocked = false
				start = newStart
			} else if opaque && j < radius {
				blocked = true
				castLight(z, origin, radius, j+1, start, left, o, vis)
				newStart = right
			}
		}
		if blocked {
			break
		}
	}
}

// updateFOV recomputes what the player sees and adds it to what they have
// explored in this room. It returns the cells seen for the first time, in
// reading order. Runs in the game loop.
func (s *Server) updateFOV(p *Player) []Pos {
	z := s.world.zoneOf(p.Room)
	p.fov = ComputeFOV(z, p.Pos, PlayerSight)
	if p.explored == nil {
		p.explored = map[string]map[Pos]bool{}
	}
	ex := p.explored[p.Room.ID]
	if ex == nil {
		ex = map[Pos]bool{}
		p.explored[p.Room.ID] = ex
	}
	var fresh []Pos
	for y := 0; y < z.H; y++ {
		for x := 0; x < z.W; x++ {
			if c := (Pos{x, y}); p.fov[c] && !ex[c] {
				ex[c] = true
				fresh = append(fresh, c)
			}
		}
	}
	return fresh
}

// visibleRows is the field of view as rows of '1' (visible) and '0'.
func visibleRows(z *Zone, fov map[Pos]bool) []string {
	rows := make([]string, z.H)
	for y := 0; y < z.H; y++ {
		b := make([]byte, z.W)
		for x := range b {
			b[x] = '0'
			if fov[Pos{x, y}] {
				b[x] = '1'
			}
		}
		rows[y] = string(b)
	}
	return rows
}
