package mud

import (
	"fmt"
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
	// TODO(lesson 18): a random number from d[0] to d[1] with s.rng.IntN
	return d[0]
}

// cmdKill starts a fight: kill <creature>. Runs in the game loop.
func cmdKill(s *Server, p *Player, cmd Command) {
	// TODO(lesson 18): MsgAttackWhat, MsgNoSuchTarget, MsgCantAttack, or
	// set p.target and send MsgYouAttack / MsgAttacks (kind KindCombat).
}

// cmdFlee runs away through a random exit.
func cmdFlee(s *Server, p *Player, cmd Command) {
	// TODO(lesson 18): MsgNotFighting / MsgFleeNowhere, or end the fight,
	// send MsgYouFlee and walk through a random exit (moveCommand(dir)(s, p, cmd)).
}

// fighting reports whether any player is fighting n. Runs in the game loop.
func (s *Server) fighting(n *NPC) bool {
	// TODO(lesson 18)
	return false
}

// combatRound lets every fighting player and their opponent hit each other
// once. Runs every tick, in the game loop.
func (s *Server) combatRound() {
	// TODO(lesson 18), see the README: for every fighting player (in name
	// order), the player hits, then – if the creature is still alive – the
	// creature hits back.
}

// defeat removes a beaten creature until it respawns.
func (s *Server) defeat(p *Player, n *NPC) {
	// TODO(lesson 18)
}

// die: the player wakes up in the start room with full health.
func (s *Server) die(p *Player) {
	// TODO(lesson 18)
}

// respawnCreatures brings defeated creatures back to where they started.
func (s *Server) respawnCreatures() {
	// TODO(lesson 18)
}

func (s *Server) sendStatusIfStructured(p *Player) {
	if p.Structured() {
		s.sendStatus(p)
	}
}
