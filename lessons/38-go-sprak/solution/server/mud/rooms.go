package mud

import (
	"fmt"
	"strings"
)

// Commands and helpers that deal with rooms. Everything here runs in the
// game loop, so no locks are needed.

// describeRoom describes the room the player is standing in.
func (s *Server) describeRoom(p *Player) string {
	r := p.Room
	text := s.roomText(p, r) // in the player's language (lesson 38)
	var b strings.Builder
	b.WriteString(text.Name + "\n")
	b.WriteString(text.Description + "\n")

	var exits []string
	for _, dir := range Directions {
		if _, ok := r.Exits[dir]; ok {
			exits = append(exits, directionNames[dir])
		}
	}
	if len(exits) == 0 {
		b.WriteString(MsgNoExits)
	} else {
		b.WriteString(fmt.Sprintf(MsgExits, strings.Join(exits, MsgListSep)))
	}

	if len(r.Items) > 0 {
		b.WriteString("\n" + fmt.Sprintf(MsgItemsHere, itemNames(r.Items)))
	}

	var others []string
	for _, other := range s.playersInRoom(r, p) {
		others = append(others, other.Name)
	}
	for _, n := range r.NPCs {
		others = append(others, n.Name)
	}
	if len(others) > 0 {
		b.WriteString("\n" + fmt.Sprintf(MsgHere, strings.Join(others, MsgListSep)))
	}
	return b.String()
}

// playersInRoom returns the players in the room except except, sorted by name.
func (s *Server) playersInRoom(r *Room, except *Player) []*Player {
	var list []*Player
	for _, name := range s.playerNames() {
		if other := s.players[name]; other.Room == r && other != except {
			list = append(list, other)
		}
	}
	return list
}

// broadcastRoom sends text of a kind to everyone in the room except except.
func (s *Server) broadcastRoom(r *Room, except *Player, kind, text string) {
	for _, other := range s.playersInRoom(r, except) {
		other.SendKind(kind, text)
	}
}

// look shows the room to the player. Graphical clients also get a map and a status.
func (s *Server) look(p *Player) {
	p.SendKind(KindRoom, s.describeRoom(p))
	if p.Structured() {
		s.sendMap(p)
		s.sendStatus(p)
		s.sendZone(p)
	}
}

// mapRadius is how many rooms in each direction the map shows.
const mapRadius = 3

func (s *Server) sendMap(p *Player) {
	m := s.world.RenderMap(p.Room, mapRadius).WithSelf()
	p.SendFrame(Frame{Type: "map", Map: &m})
}

func (s *Server) sendStatus(p *Player) {
	var target *Target
	if n := p.target; n != nil {
		target = &Target{Name: n.Name, HP: max(n.hp, 0), MaxHP: n.HP}
	}
	p.SendFrame(Frame{Type: "status", Status: &Status{
		Name:    p.Name,
		Room:    s.roomText(p, p.Room).Name,
		Terrain: p.Room.Terrain,
		HP:      p.HP,
		MaxHP:   p.MaxHP,
		Online:  len(s.players),
		Target:  target,
	}})
}

func cmdLook(s *Server, p *Player, cmd Command) {
	s.look(p)
}

func cmdMap(s *Server, p *Player, cmd Command) {
	s.sendMap(p)
}

// travel takes the player through the exit in direction dir to the next room,
// where they appear at the matching opening of its zone.
func (s *Server) travel(p *Player, dir string) {
	to := s.world.Room(p.Room.Exits[dir])
	if to == nil {
		p.SendKind(KindError, MsgCantGo)
		return
	}
	if e := (&Event{Kind: EventEnter, Player: p, Room: p.Room, To: to, Dir: dir}); !s.emit(e) {
		e.refuse() // e.g. "The door is locked."
		return
	}
	s.broadcastRoom(p.Room, p, KindMove, fmt.Sprintf(MsgLeaves, p.Name, directionTravel[dir]))
	p.Room = to
	p.Pos = s.world.zoneOf(to).arrival(opposite[dir])
	s.broadcastRoom(to, p, KindMove, fmt.Sprintf(MsgArrives, p.Name))
	s.look(p)
}

// stepCommand creates a command that takes one step in direction dir.
// It returns a closure that "remembers" dir.
func stepCommand(dir string) CommandFunc {
	return func(s *Server, p *Player, cmd Command) { s.step(p, dir) }
}

// step moves the player one cell. Out through an opening in the edge of the
// zone it becomes travel to the next room.
func (s *Server) step(p *Player, dir string) {
	z := s.world.zoneOf(p.Room)
	to := p.Pos.Add(stepDelta[dir])
	if !z.In(to) {
		if _, ok := p.Room.Exits[dir]; ok {
			s.travel(p, dir)
			return
		}
		p.SendKind(KindError, MsgBlocked)
		return
	}
	if !Walkable(z.At(to)) {
		p.SendKind(KindError, MsgBlocked)
		return
	}
	p.Pos = to
	if !p.Structured() {
		p.SendFrame(Frame{Type: "zone", Zone: s.zoneData(p)}) // telnet: draw the zone again
	}
}

// stairsCommand creates "up" or "down": they work when you stand on stairs.
func stairsCommand(dir string) CommandFunc {
	return func(s *Server, p *Player, cmd Command) {
		want := byte(TileStairsUp)
		if dir == "down" {
			want = TileStairsDown
		}
		if s.world.zoneOf(p.Room).At(p.Pos) != want {
			p.SendKind(KindError, MsgNoStairs)
			return
		}
		s.travel(p, dir)
	}
}

// cmdGo travels straight to the next room: "go north".
func cmdGo(s *Server, p *Player, cmd Command) {
	dir := cmd.Rest
	if full, ok := aliases[dir]; ok {
		dir = full
	}
	if dir == "" {
		p.SendKind(KindError, MsgGoWhere)
		return
	}
	s.travel(p, dir)
}

// cmdZone sends the zone again (as text in telnet).
func cmdZone(s *Server, p *Player, cmd Command) {
	if p.Structured() {
		s.sendZone(p)
		return
	}
	p.SendFrame(Frame{Type: "zone", Zone: s.zoneData(p)})
}

// taken is the cells occupied by creatures and players in a room.
func (w *World) taken(r *Room) map[Pos]bool {
	t := map[Pos]bool{}
	for _, n := range r.NPCs {
		t[n.Pos] = true
	}
	return t
}

// placeItem puts an item in a room, on a free cell near the middle.
func (w *World) placeItem(r *Room, it *Item) {
	used := map[Pos]bool{}
	for _, other := range r.Items {
		used[other.Pos] = true
	}
	z := w.zoneOf(r)
	it.Pos = z.freeNear(z.Center(), used)
	r.Items = append(r.Items, it)
}

func cmdSay(s *Server, p *Player, cmd Command) {
	if cmd.Rest == "" {
		p.SendKind(KindError, MsgSayWhat)
		return
	}
	e := &Event{Kind: EventSay, Player: p, Room: p.Room, Text: cmd.Rest}
	if !s.emit(e) {
		e.refuse()
		return
	}
	p.SendKind(KindSay, fmt.Sprintf(MsgYouSay, e.Text))
	s.broadcastRoom(p.Room, p, KindSay, fmt.Sprintf(MsgSays, p.Name, e.Text))
}

func cmdShout(s *Server, p *Player, cmd Command) {
	if cmd.Rest == "" {
		p.SendKind(KindError, MsgShoutWhat)
		return
	}
	p.SendKind(KindSay, fmt.Sprintf(MsgYouShout, cmd.Rest))
	s.broadcast(p, KindSay, fmt.Sprintf(MsgShouts, p.Name, cmd.Rest))
}
