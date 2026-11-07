package mud

import (
	"sort"
	"strings"
)

// Entity is something with a position in a zone: a player, a creature or an item.
type Entity struct {
	ID    string `json:"id"`   // "p:Anna", "n:cat", "i:sword"
	Kind  string `json:"kind"` // "player", "creature" or "item"
	Name  string `json:"name"`
	Glyph string `json:"glyph"` // how a text client draws it
	X     int    `json:"x"`
	Y     int    `json:"y"`
}

// ZoneData is a whole zone: its tiles and what is in it.
type ZoneData struct {
	Room     string            `json:"room"` // room ID
	Width    int               `json:"width"`
	Height   int               `json:"height"`
	Rows     []string          `json:"rows"`
	Legend   map[string]string `json:"legend"`
	Entities []Entity          `json:"entities"`
	Self     string            `json:"self"` // the receiver's own entity ID
}

// DeltaData is what changed in a zone since the last frame – much smaller
// than sending the whole zone again every tick.
type DeltaData struct {
	Room    string   `json:"room"`
	Added   []Entity `json:"added,omitempty"`
	Moved   []Entity `json:"moved,omitempty"`
	Removed []string `json:"removed,omitempty"`
}

// glyphOf is the first character of s (the default glyph of a creature).
func glyphOf(s, fallback string) string {
	for _, r := range s {
		return string(r)
	}
	return fallback
}

// entities lists everything in a room, sorted by ID. Runs in the game loop.
func (s *Server) entities(r *Room) []Entity {
	// TODO(lesson 23): players ("p:"+name, glyph "@"), creatures ("n:"+id,
	// glyph n.Glyph or glyphOf(n.ID, "c")) and items ("i:"+id, glyph it.Glyph or "*").
	return nil
}

// zoneData is the zone the player is in, as the player sees it.
func (s *Server) zoneData(p *Player) *ZoneData {
	// TODO(lesson 23): rows, a legend with the codes that occur, entities, self.
	return &ZoneData{Room: p.Room.ID}
}

// sendZone sends the whole zone and remembers what the player now knows.
func (s *Server) sendZone(p *Player) {
	// TODO(lesson 23): send the zone frame and remember it in p.seenRoom / p.seen.
}

// syncAll sends every graphical client what changed in its zone. It runs
// after every command and every tick.
func (s *Server) syncAll() {
	// TODO(lesson 23): syncPlayer for every structured player.
}

// syncPlayer compares what the player has been told with how things are now
// and sends the difference as a delta frame (or a whole zone in a new room).
func (s *Server) syncPlayer(p *Player) {
	// TODO(lesson 23): a new room -> sendZone. Otherwise compare p.seen with
	// s.entities(p.Room) and send a delta with added/moved/removed (if anything changed).
}

// renderZone draws a zone with its entities as text, for telnet.
func renderZone(zd *ZoneData) string {
	grid := make([][]rune, len(zd.Rows))
	for y, row := range zd.Rows {
		grid[y] = []rune(row)
	}
	// Items first, then creatures, then players – the most important on top.
	order := map[string]int{"item": 0, "creature": 1, "player": 2}
	ents := append([]Entity(nil), zd.Entities...)
	sort.SliceStable(ents, func(i, j int) bool { return order[ents[i].Kind] < order[ents[j].Kind] })
	for _, e := range ents {
		if e.Y >= 0 && e.Y < len(grid) && e.X >= 0 && e.X < len(grid[e.Y]) {
			grid[e.Y][e.X] = []rune(e.Glyph)[0]
		}
	}
	lines := make([]string, len(grid))
	for y, row := range grid {
		lines[y] = string(row)
	}
	return strings.Join(lines, "\r\n") + "\r\n"
}
