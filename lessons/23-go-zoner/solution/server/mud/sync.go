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
	var list []Entity
	for _, p := range s.playersInRoom(r, nil) {
		list = append(list, Entity{ID: "p:" + p.Name, Kind: "player", Name: p.Name, Glyph: "@", X: p.Pos.X, Y: p.Pos.Y})
	}
	for _, n := range r.NPCs {
		g := n.Glyph
		if g == "" {
			g = glyphOf(n.ID, "c")
		}
		list = append(list, Entity{ID: "n:" + n.ID, Kind: "creature", Name: n.Name, Glyph: g, X: n.Pos.X, Y: n.Pos.Y})
	}
	for _, it := range r.Items {
		g := it.Glyph
		if g == "" {
			g = "*"
		}
		list = append(list, Entity{ID: "i:" + it.ID, Kind: "item", Name: it.Name, Glyph: g, X: it.Pos.X, Y: it.Pos.Y})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	return list
}

// zoneData is the zone the player is in, as the player sees it.
func (s *Server) zoneData(p *Player) *ZoneData {
	z := s.world.zoneOf(p.Room)
	legend := map[string]string{}
	for _, row := range z.Tiles {
		for _, t := range row {
			legend[string(t)] = tileMeanings[t]
		}
	}
	return &ZoneData{
		Room: p.Room.ID, Width: z.W, Height: z.H, Rows: z.Rows(), Legend: legend,
		Entities: s.entities(p.Room), Self: "p:" + p.Name,
	}
}

// sendZone sends the whole zone and remembers what the player now knows.
func (s *Server) sendZone(p *Player) {
	zd := s.zoneData(p)
	p.seenRoom = p.Room.ID
	p.seen = make(map[string]Entity, len(zd.Entities))
	for _, e := range zd.Entities {
		p.seen[e.ID] = e
	}
	p.SendFrame(Frame{Type: "zone", Zone: zd})
}

// syncAll sends every graphical client what changed in its zone. It runs
// after every command and every tick.
func (s *Server) syncAll() {
	for _, p := range s.players {
		if p.Structured() && p.Room != nil {
			s.syncPlayer(p)
		}
	}
}

// syncPlayer compares what the player has been told with how things are now
// and sends the difference as a delta frame (or a whole zone in a new room).
func (s *Server) syncPlayer(p *Player) {
	if p.seenRoom != p.Room.ID {
		s.sendZone(p)
		return
	}
	now := s.entities(p.Room)
	d := DeltaData{Room: p.Room.ID}
	current := make(map[string]Entity, len(now))
	for _, e := range now {
		current[e.ID] = e
		old, known := p.seen[e.ID]
		switch {
		case !known:
			d.Added = append(d.Added, e)
		case old != e:
			d.Moved = append(d.Moved, e)
		}
	}
	for id := range p.seen {
		if _, still := current[id]; !still {
			d.Removed = append(d.Removed, id)
		}
	}
	sort.Strings(d.Removed)
	p.seen = current
	if len(d.Added)+len(d.Moved)+len(d.Removed) > 0 {
		p.SendFrame(Frame{Type: "delta", Delta: &d})
	}
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
