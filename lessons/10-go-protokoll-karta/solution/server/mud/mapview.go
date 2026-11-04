package mud

import "strings"

// Koder för kartan. De är valda så att kartan går att läsa även som ren
// ASCII i telnet.
var terrainCodes = map[string]byte{
	"town": '#', "inside": '+', "road": '.', "field": '"',
	"forest": 'T', "water": '~', "mountain": '^', "cave": 'o',
}

// mapLegend översätter varje kod till ett namn.
var mapLegend = map[string]string{
	"#": "town", "+": "inside", ".": "road", "\"": "field",
	"T": "forest", "~": "water", "^": "mountain", "o": "cave",
	"<": "stairs_up", ">": "stairs_down", "@": "self",
	"-": "path_ew", "|": "path_ns",
}

// RenderMap ritar rummen inom radius steg från center (på samma våning).
//
// Rummen ligger på varannan rad och kolumn; mellan dem ritas utgångar:
//
//   - krogen
//     |          utgång north/south
//     #-+        torget - smedjan
//
// Ett rum med trappa upp ritas '<' och med trappa ner '>'.
// Kartan blir (4*radius+1) tecken bred och hög, med center i mitten.
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

// WithSelf returnerar en kopia av kartan med '@' (spelaren) i mitten.
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
