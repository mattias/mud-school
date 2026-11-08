package mud

import (
	"fmt"
	"sort"
)

// Combat settings.
const (
	PlayerMaxHP    = 20
	DefaultRespawn = 50 // ticks before a defeated creature comes back
)

// PlayerDamage is how much a player's hit does: a random number from min to max.
var PlayerDamage = [2]int{1, 4}

// roll returns a random number from d[0] to d[1] (both included).
func (s *Server) roll(d [2]int) int {
	if d[1] <= d[0] {
		return d[0]
	}
	return d[0] + s.rng.IntN(d[1]-d[0]+1)
}

// cmdKill starts a fight: kill <creature>. Runs in the game loop.
func cmdKill(s *Server, p *Player, cmd Command) {
	if cmd.Rest == "" {
		p.SendKind(KindError, MsgAttackWhat)
		return
	}
	n, ok := FindMatch(p.Room.NPCs, cmd.Rest)
	if !ok {
		p.SendKind(KindError, fmt.Sprintf(MsgNoSuchTarget, cmd.Rest))
		return
	}
	if !n.Attackable {
		p.SendKind(KindError, fmt.Sprintf(MsgCantAttack, n.Name))
		return
	}
	p.target = n
	p.SendKind(KindCombat, fmt.Sprintf(MsgYouAttack, n.Name))
	s.broadcastRoom(p.Room, p, KindCombat, fmt.Sprintf(MsgAttacks, p.Name, n.Name))
}

// cmdFlee runs away through a random exit.
func cmdFlee(s *Server, p *Player, cmd Command) {
	if p.target == nil {
		p.SendKind(KindError, MsgNotFighting)
		return
	}
	var dirs []string
	for _, dir := range Directions {
		if _, ok := p.Room.Exits[dir]; ok {
			dirs = append(dirs, dir)
		}
	}
	if len(dirs) == 0 {
		p.SendKind(KindError, MsgFleeNowhere)
		return
	}
	dir := dirs[s.rng.IntN(len(dirs))]
	p.target = nil
	p.SendKind(KindCombat, fmt.Sprintf(MsgYouFlee, directionTravel[dir]))
	s.travel(p, dir)
}

// fighting reports whether any player is fighting n. Runs in the game loop.
func (s *Server) fighting(n *NPC) bool {
	for _, p := range s.players {
		if p.target == n {
			return true
		}
	}
	return false
}

// combatRound lets every fighting player and their opponent hit each other
// once. Runs every tick, in the game loop.
func (s *Server) combatRound() {
	names := s.playerNames() // a fixed order: the same seed gives the same fight
	for _, name := range names {
		p := s.players[name]
		n := p.target
		if n == nil {
			continue
		}
		if n.dead || n.Location != p.Room {
			p.target = nil
			continue
		}
		dmg := s.roll(PlayerDamage)
		n.hp -= dmg
		p.SendKind(KindCombat, fmt.Sprintf(MsgYouHit, n.Name, dmg))
		if n.hp <= 0 {
			s.defeat(p, n)
			s.sendStatusIfStructured(p)
			continue
		}
		dmg = s.roll(n.Damage)
		p.HP -= dmg
		p.SendKind(KindCombat, fmt.Sprintf(MsgHitsYou, n.Name, dmg))
		if p.HP <= 0 {
			s.die(p)
		}
		s.sendStatusIfStructured(p)
	}
}

// defeat removes a beaten creature until it respawns.
func (s *Server) defeat(p *Player, n *NPC) {
	p.SendKind(KindCombat, fmt.Sprintf(MsgYouKill, n.Name))
	s.broadcastRoom(p.Room, p, KindCombat, fmt.Sprintf(MsgKills, p.Name, n.Name))
	for _, other := range s.players {
		if other.target == n {
			other.target = nil
		}
	}
	n.dead = true
	n.Location.NPCs = remove(n.Location.NPCs, n)
	n.respawnAt = s.ticks + int64(s.respawn)
	s.emit(&Event{Kind: EventDeath, Player: p, NPC: n, Room: n.Location}) // lesson 34
}

// die: the player wakes up in the start room with full health.
func (s *Server) die(p *Player) {
	s.emit(&Event{Kind: EventDeath, Player: p, Room: p.Room}) // lesson 34
	start := s.world.Room(s.world.Start)
	s.broadcastRoom(p.Room, p, KindCombat, fmt.Sprintf(MsgDies, p.Name))
	p.target = nil
	p.Room = start
	p.Pos = s.world.zoneOf(start).freeNear(s.world.zoneOf(start).Center(), nil)
	p.HP = p.MaxHP
	p.SendKind(KindCombat, fmt.Sprintf(MsgYouDie, start.Name))
	s.broadcastRoom(start, p, KindMove, fmt.Sprintf(MsgArrives, p.Name))
	s.look(p)
}

// respawnCreatures brings defeated creatures back to where they started.
func (s *Server) respawnCreatures() {
	var due []*NPC
	for _, n := range s.world.NPCs {
		if n.dead && s.ticks >= n.respawnAt {
			due = append(due, n)
		}
	}
	sort.Slice(due, func(i, j int) bool { return due[i].ID < due[j].ID })
	for _, n := range due {
		n.dead = false
		n.hp = n.HP
		n.Location = n.origin
		n.Pos = n.home
		n.Location.NPCs = append(n.Location.NPCs, n)
		s.broadcastRoom(n.Location, nil, KindMove, fmt.Sprintf(MsgAppears, n.Name))
	}
}

func (s *Server) sendStatusIfStructured(p *Player) {
	if p.Structured() {
		s.sendStatus(p)
	}
}
