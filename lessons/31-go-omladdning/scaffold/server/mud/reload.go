package mud

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// Reloading the world while the server runs (lesson 31): the world file is
// watched, loaded and validated OUTSIDE the game loop, and swapped in one
// go INSIDE it – nobody ever sees half a world.

// isAdmin reports whether the player may use the @ commands.
func (s *Server) isAdmin(p *Player) bool { return s.admins[p.Name] }

// adminOnly wraps a command so that only administrators can use it.
func adminOnly(fn CommandFunc) CommandFunc {
	return func(s *Server, p *Player, cmd Command) {
		if !s.isAdmin(p) {
			p.SendKind(KindError, MsgNotAdmin)
			return
		}
		fn(s, p, cmd)
	}
}

// notifyAdmins sends a message to every administrator who is logged in.
func (s *Server) notifyAdmins(kind, text string) {
	for _, p := range s.players {
		if s.isAdmin(p) {
			p.SendKind(kind, text)
		}
	}
}

// cmdReload: "@reload" – load the world file again now.
func cmdReload(s *Server, p *Player, cmd Command) {
	if s.worldPath == "" {
		p.SendKind(KindError, MsgNoWorldFile)
		return
	}
	go s.reloadFromFile() // reading the disk must not stop the game loop
}

// reloadFromFile loads and validates the world file, then swaps it in – or
// tells the administrators what is wrong. Runs outside the game loop.
func (s *Server) reloadFromFile() {
	// TODO(lesson 31): LoadWorldFile(s.worldPath) HERE (outside the loop), then
	// in the loop (s.do): on an error log it and notifyAdmins(MsgReloadFailed);
	// otherwise swapWorld.
	_ = log.Printf
}

// watchWorld checks the world file every s.reloadEvery and reloads it when
// it has changed. It stops when the server stops.
func (s *Server) watchWorld() {
	// TODO(lesson 31): a time.Ticker every s.reloadEvery; os.Stat the file and
	// reload when ModTime or Size changed. Return when s.stop is closed.
	_, _ = os.Stat, time.Now
}

// swapWorld replaces the world. Runs in the game loop. Players keep their
// room (or go to the start room if it is gone), their position if it is
// still walkable, their health and their things.
func (s *Server) swapWorld(nw *World) {
	// TODO(lesson 31): s.world = nw, then for every player: the room with the
	// same ID (or the start room + MsgRoomGone), a walkable position, no
	// target, the new world's version of every item (s.takeItem), and forget
	// explored/seenRoom. Finally MsgReloaded + look for everyone.
	_ = nw
	_ = fmt.Sprintf
}

// normalizeAdmins turns the admin names into a set of normalized names.
func normalizeAdmins(names []string) map[string]bool {
	set := map[string]bool{}
	for _, n := range names {
		if n = strings.TrimSpace(n); n != "" {
			set[NormalizeName(n)] = true
		}
	}
	return set
}
