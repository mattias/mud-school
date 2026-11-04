package mud

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
	// TODO(lesson 10)
	return MapData{}
}

// WithSelf returns a copy of the map with '@' (the player) in the middle.
func (m MapData) WithSelf() MapData {
	// TODO(lesson 10). Careful: m.Rows and m.Legend are a slice and a map –
	// if you don't copy them you change the original!
	return m
}
