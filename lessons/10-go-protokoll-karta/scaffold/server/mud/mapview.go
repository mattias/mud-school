package mud

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
// Legend ska innehålla de koder som faktiskt används på kartan.
func (w *World) RenderMap(center *Room, radius int) MapData {
	// TODO(lektion 10)
	return MapData{}
}

// WithSelf returnerar en kopia av kartan med '@' (spelaren) i mitten.
func (m MapData) WithSelf() MapData {
	// TODO(lektion 10). Var försiktig: m.Rows och m.Legend är en slice och
	// en map – kopierar du inte dem ändrar du originalet!
	return m
}
