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
	nw, err := LoadWorldFile(s.worldPath)
	s.do(func() {
		if err != nil {
			log.Printf("reload: %v", err)
			s.notifyAdmins(KindError, fmt.Sprintf(MsgReloadFailed, err))
			return
		}
		s.swapWorld(nw)
	})
}

// watchWorld checks the world file every s.reloadEvery and reloads it when
// it has changed. It stops when the server stops.
func (s *Server) watchWorld() {
	stat := func() (time.Time, int64) {
		fi, err := os.Stat(s.worldPath)
		if err != nil {
			return time.Time{}, -1
		}
		return fi.ModTime(), fi.Size()
	}
	lastMod, lastSize := stat()
	t := time.NewTicker(s.reloadEvery)
	defer t.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-t.C:
			s.saveMu.Lock() // not in the middle of our own save (lesson 36)
			mod, size := stat()
			own, _ := s.ownWrite.Load().(fileStamp)
			s.saveMu.Unlock()
			if size < 0 || (mod.Equal(lastMod) && size == lastSize) {
				continue
			}
			lastMod, lastSize = mod, size
			if own.mod.Equal(mod) && own.size == size {
				continue // we wrote it ourselves with @dig & co (lesson 36): nothing to reload
			}
			s.reloadFromFile()
		}
	}
}

// swapWorld replaces the world. Runs in the game loop. Players keep their
// room (or go to the start room if it is gone), their position if it is
// still walkable, their health and their things.
func (s *Server) swapWorld(nw *World) {
	s.world = nw
	for _, p := range s.players {
		room := nw.Room(p.Room.ID)
		if room == nil {
			room = nw.Room(nw.Start)
			p.SendKind(KindMove, MsgRoomGone)
		}
		z := nw.zoneOf(room)
		if room.ID != p.Room.ID || !Walkable(z.At(p.Pos)) {
			p.Pos = z.freeNear(z.Center(), nil)
		}
		p.Room = room
		p.target = nil // the creatures are new
		// The things they carry: the new world's version, taken out of its room.
		for i, it := range p.Inventory {
			if fresh := s.takeItem(it.ID); fresh != nil {
				p.Inventory[i] = fresh
			}
		}
		// The zone may look different now: forget what was explored and sent.
		p.explored = nil
		p.seenRoom = ""
	}
	log.Printf("the world was reloaded: %d rooms", len(nw.Rooms))
	for _, p := range s.players {
		p.SendKind(KindInfo, MsgReloaded)
		s.look(p)
	}
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
