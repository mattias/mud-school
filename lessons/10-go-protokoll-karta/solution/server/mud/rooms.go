package mud

import (
	"fmt"
	"strings"
)

// Commands and helpers that deal with rooms.
// Every function whose name ends in "Locked" assumes that s.mu is already held.

// describeRoomLocked describes the room the player is standing in.
func (s *Server) describeRoomLocked(p *Player) string {
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

	var others []string
	for _, other := range s.playersInRoomLocked(r, p) {
		others = append(others, other.Name)
	}
	if len(others) > 0 {
		b.WriteString("\n" + fmt.Sprintf(MsgHere, strings.Join(others, MsgListSep)))
	}
	return b.String()
}

// playersInRoomLocked returns the players in the room except except, sorted by name.
func (s *Server) playersInRoomLocked(r *Room, except *Player) []*Player {
	var list []*Player
	for _, name := range s.playerNamesLocked() {
		if other := s.players[name]; other.Room == r && other != except {
			list = append(list, other)
		}
	}
	return list
}

// broadcastRoomLocked sends text to everyone in the room except except.
func (s *Server) broadcastRoomLocked(r *Room, except *Player, text string) {
	for _, other := range s.playersInRoomLocked(r, except) {
		other.Send(text)
	}
}

// lookLocked shows the room to the player. Graphical clients also get a map and a status.
func (s *Server) lookLocked(p *Player) {
	p.Send(s.describeRoomLocked(p))
	if p.Structured() {
		s.sendMapLocked(p)
		s.sendStatusLocked(p)
	}
}

// mapRadius is how many rooms in each direction the map shows.
const mapRadius = 3

func (s *Server) sendMapLocked(p *Player) {
	m := s.world.RenderMap(p.Room, mapRadius).WithSelf()
	p.SendFrame(Frame{Type: "map", Map: &m})
}

func (s *Server) sendStatusLocked(p *Player) {
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
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lookLocked(p)
}

func cmdMap(s *Server, p *Player, cmd Command) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sendMapLocked(p)
}

// moveCommand creates a command that walks in direction dir.
// It returns a closure that "remembers" dir.
func moveCommand(dir string) CommandFunc {
	return func(s *Server, p *Player, cmd Command) {
		s.mu.Lock()
		defer s.mu.Unlock()
		to := s.world.Room(p.Room.Exits[dir])
		if to == nil {
			p.Send(MsgCantGo)
			return
		}
		s.broadcastRoomLocked(p.Room, p, fmt.Sprintf(MsgLeaves, p.Name, directionTravel[dir]))
		p.Room = to
		s.broadcastRoomLocked(to, p, fmt.Sprintf(MsgArrives, p.Name))
		s.lookLocked(p)
	}
}

func cmdSay(s *Server, p *Player, cmd Command) {
	if cmd.Rest == "" {
		p.Send(MsgSayWhat)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p.Sendf(MsgYouSay, cmd.Rest)
	s.broadcastRoomLocked(p.Room, p, fmt.Sprintf(MsgSays, p.Name, cmd.Rest))
}

func cmdShout(s *Server, p *Player, cmd Command) {
	if cmd.Rest == "" {
		p.Send(MsgShoutWhat)
		return
	}
	p.Sendf(MsgYouShout, cmd.Rest)
	s.broadcast(p, fmt.Sprintf(MsgShouts, p.Name, cmd.Rest))
}
