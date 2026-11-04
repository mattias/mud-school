package mud

import "strings"

// Kommandon och hjälpfunktioner som har med rum att göra.
// Alla funktioner med "Locked" i namnet förutsätter att s.mu redan är låst.

// describeRoomLocked beskriver rummet som spelaren står i.
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
		b.WriteString("Utgångar: inga")
	} else {
		b.WriteString("Utgångar: " + strings.Join(exits, ", "))
	}

	var others []string
	for _, other := range s.playersInRoomLocked(r, p) {
		others = append(others, other.Name)
	}
	if len(others) > 0 {
		b.WriteString("\nHär finns: " + strings.Join(others, ", "))
	}
	return b.String()
}

// playersInRoomLocked returnerar spelarna i rummet utom except, sorterade på namn.
func (s *Server) playersInRoomLocked(r *Room, except *Player) []*Player {
	var list []*Player
	for _, name := range s.playerNamesLocked() {
		if other := s.players[name]; other.Room == r && other != except {
			list = append(list, other)
		}
	}
	return list
}

// broadcastRoomLocked skickar text till alla i rummet utom except.
func (s *Server) broadcastRoomLocked(r *Room, except *Player, text string) {
	for _, other := range s.playersInRoomLocked(r, except) {
		other.Send(text)
	}
}

// lookLocked visar rummet för spelaren. Grafiska klienter får också karta och status.
func (s *Server) lookLocked(p *Player) {
	p.Send(s.describeRoomLocked(p))
	if p.Structured() {
		s.sendMapLocked(p)
		s.sendStatusLocked(p)
	}
}

// mapRadius är hur många rum åt varje håll kartan visar.
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
		HP:      10, // strid kommer i fas 3!
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

// moveCommand skapar ett kommando för att gå i riktningen dir.
// Funktionen returnerar en closure som "kommer ihåg" dir.
func moveCommand(dir string) CommandFunc {
	return func(s *Server, p *Player, cmd Command) {
		s.mu.Lock()
		defer s.mu.Unlock()
		to := s.world.Room(p.Room.Exits[dir])
		if to == nil {
			p.Send("Du kan inte gå åt det hållet.")
			return
		}
		s.broadcastRoomLocked(p.Room, p, p.Name+" går "+directionTravel[dir]+".")
		p.Room = to
		s.broadcastRoomLocked(to, p, p.Name+" kommer in.")
		s.lookLocked(p)
	}
}

func cmdSay(s *Server, p *Player, cmd Command) {
	if cmd.Rest == "" {
		p.Send("Säg vad?")
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p.Sendf("Du säger: %s", cmd.Rest)
	s.broadcastRoomLocked(p.Room, p, p.Name+" säger: "+cmd.Rest)
}

func cmdShout(s *Server, p *Player, cmd Command) {
	if cmd.Rest == "" {
		p.Send("Ropa vad?")
		return
	}
	p.Sendf("Du ropar: %s", cmd.Rest)
	s.broadcast(p, p.Name+" ropar: "+cmd.Rest)
}
