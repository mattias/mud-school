package mud

import "strings"

// Map codes. They are chosen so that the map is readable as plain ASCII in telnet.
var terrainCodes = map[string]byte{
	"town": '#', "inside": '+', "road": '.', "field": '"',
	"forest": 'T', "water": '~', "mountain": '^', "cave": 'o',
}

// mapLegend translates every code to a meaning.
var mapLegend = map[string]string{
	"#": "town", "+": "inside", ".": "road", "\"": "field",
	"T": "forest", "~": "water", "^": "mountain", "o": "cave",
	"<": "stairs_up", ">": "stairs_down", "@": "self",
	"-": "path_ew", "|": "path_ns",
}

// RenderMap draws the rooms within radius steps of center (on the same floor).
//
// Rooms are placed on every other row and column; exits are drawn between them:
//
//   - tavern
//     |          exit north/south
//     #-+        square - smithy
//
// A room with stairs up is drawn '<', with stairs down '>'.
// The map is (4*radius+1) characters wide and high, with center in the middle.
// Legend contains the codes that are actually used on the map.
func (w *World) RenderMap(center *Room, radius int) MapData {
	size := 4*radius + 1
	grid := make([][]byte, size)
	for i := range grid {
		grid[i] = []byte(strings.Repeat(" ", size))
	}
	for _, r := range w.Rooms {
		dx, dy := r.X-center.X, r.Y-center.Y
		if r.Z != center.Z || dx < -radius || dx > radius || dy < -radius || dy > radius {
			continue
		}
		col, row := 2*(dx+radius), 2*(dy+radius)
		code := terrainCodes[r.Terrain]
		if _, ok := r.Exits["down"]; ok {
			code = '>'
		}
		if _, ok := r.Exits["up"]; ok {
			code = '<'
		}
		grid[row][col] = code
		if _, ok := r.Exits["east"]; ok && col+1 < size {
			grid[row][col+1] = '-'
		}
		if _, ok := r.Exits["south"]; ok && row+1 < size {
			grid[row+1][col] = '|'
		}
	}
	rows := make([]string, size)
	used := make(map[string]string)
	for i, line := range grid {
		rows[i] = string(line)
		for _, ch := range line {
			if name, ok := mapLegend[string(ch)]; ok {
				used[string(ch)] = name
			}
		}
	}
	return MapData{Width: size, Height: size, Rows: rows, Legend: used}
}

// WithSelf returns a copy of the map with '@' (the player) in the middle.
func (m MapData) WithSelf() MapData {
	mid := m.Height / 2
	rows := append([]string(nil), m.Rows...)
	line := []byte(rows[mid])
	line[m.Width/2] = '@'
	rows[mid] = string(line)

	legend := make(map[string]string, len(m.Legend)+1)
	for k, v := range m.Legend {
		legend[k] = v
	}
	legend["@"] = "self"
	m.Rows, m.Legend = rows, legend
	return m
}
