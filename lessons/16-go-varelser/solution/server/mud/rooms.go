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
	var b strings.Builder
	b.WriteString(r.Name + "\n")
	b.WriteString(r.Description + "\n")

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
	}
}

// mapRadius is how many rooms in each direction the map shows.
const mapRadius = 3

func (s *Server) sendMap(p *Player) {
	m := s.world.RenderMap(p.Room, mapRadius).WithSelf()
	p.SendFrame(Frame{Type: "map", Map: &m})
}

func (s *Server) sendStatus(p *Player) {
	p.SendFrame(Frame{Type: "status", Status: &Status{
		Name:    p.Name,
		Room:    p.Room.Name,
		Terrain: p.Room.Terrain,
		HP:      10, // combat arrives in phase 3!
		MaxHP:   10,
		Online:  len(s.players),
	}})
}

func cmdLook(s *Server, p *Player, cmd Command) {
	s.look(p)
}

func cmdMap(s *Server, p *Player, cmd Command) {
	s.sendMap(p)
}

// moveCommand creates a command that walks in direction dir.
// It returns a closure that "remembers" dir.
func moveCommand(dir string) CommandFunc {
	return func(s *Server, p *Player, cmd Command) {
		to := s.world.Room(p.Room.Exits[dir])
		if to == nil {
			p.SendKind(KindError, MsgCantGo)
			return
		}
		s.broadcastRoom(p.Room, p, KindMove, fmt.Sprintf(MsgLeaves, p.Name, directionTravel[dir]))
		p.Room = to
		s.broadcastRoom(to, p, KindMove, fmt.Sprintf(MsgArrives, p.Name))
		s.look(p)
	}
}

func cmdSay(s *Server, p *Player, cmd Command) {
	if cmd.Rest == "" {
		p.SendKind(KindError, MsgSayWhat)
		return
	}
	p.SendKind(KindSay, fmt.Sprintf(MsgYouSay, cmd.Rest))
	s.broadcastRoom(p.Room, p, KindSay, fmt.Sprintf(MsgSays, p.Name, cmd.Rest))
}

func cmdShout(s *Server, p *Player, cmd Command) {
	if cmd.Rest == "" {
		p.SendKind(KindError, MsgShoutWhat)
		return
	}
	p.SendKind(KindSay, fmt.Sprintf(MsgYouShout, cmd.Rest))
	s.broadcast(p, KindSay, fmt.Sprintf(MsgShouts, p.Name, cmd.Rest))
}
