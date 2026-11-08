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
	// TODO(lesson 26): recursive shadowcasting – see the README and HINTS.
}

// updateFOV recomputes what the player sees and adds it to what they have
// explored in this room. It returns the cells seen for the first time, in
// reading order. Runs in the game loop.
func (s *Server) updateFOV(p *Player) []Pos {
	// TODO(lesson 26): p.fov = ComputeFOV(...); add every visible cell to
	// p.explored[p.Room.ID] and return the ones that were new, in reading order.
	return nil
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
