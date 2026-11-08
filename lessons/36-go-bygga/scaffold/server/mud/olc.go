package mud

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

// Building in the game (lesson 36): administrators change the world from
// inside it – "online creation". Every change is validated (lesson 9), saved
// to the world file, and visible to everyone at once.

// fileStamp is a file's modification time and size, to recognise our own writes.
type fileStamp struct {
	mod  time.Time
	size int64
}

// slug turns a room name into an ID: "Dark cave" -> "dark-cave".
func slug(name string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(name) {
		if r < 128 && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
			b.WriteRune(r)
			dash = false
		} else if b.Len() > 0 && !dash {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

// newRoomID makes a free room ID from a name.
func (w *World) newRoomID(name string) string {
	base := slug(name)
	if base == "" {
		base = "room"
	}
	id := base
	for n := 2; w.Room(id) != nil; n++ {
		id = fmt.Sprintf("%s-%d", base, n)
	}
	return id
}

// change applies fn to the world, validates, and undoes it with undo if the
// world is no longer valid. It returns the validation error, if any.
func (s *Server) change(fn, undo func()) error {
	fn()
	if err := s.world.Validate(); err != nil {
		undo()
		return err
	}
	return nil
}

// rezone forgets a room's zone (its exits changed), and makes sure everything
// in it still stands on a walkable cell.
func (s *Server) rezone(r *Room) {
	r.zone = nil
	z := s.world.zoneOf(r)
	fix := func(p *Pos) {
		if !Walkable(z.At(*p)) {
			*p = z.freeNear(z.Center(), nil)
		}
	}
	for _, p := range s.playersInRoom(r, nil) {
		fix(&p.Pos)
		if p.explored != nil {
			delete(p.explored, r.ID)
		}
		p.seenRoom = "" // send the zone again
	}
	for _, n := range r.NPCs {
		fix(&n.Pos)
	}
	for _, it := range r.Items {
		fix(&it.Pos)
	}
}

// cmdDig: "@dig north Dark cave" – a new room in that direction, with exits both ways.
func cmdDig(s *Server, p *Player, cmd Command) {
	// TODO(lesson 36): parse "<direction> <name>" (MsgDigUsage); the exit must be
	// free (MsgExitExists); a new Room with newRoomID, MsgNewRoomDesc, the same
	// terrain, the coordinates one step away and an exit back; add it and the
	// exit with s.change (MsgBuildInvalid on an error); rezone; MsgDug; saveWorld.
	_ = directionDelta
}

// cmdDesc: "@desc Water drips from the ceiling." – the room's description.
func cmdDesc(s *Server, p *Player, cmd Command) {
	if cmd.Rest == "" {
		p.SendKind(KindError, MsgDescUsage)
		return
	}
	p.Room.Description = cmd.Rest
	p.SendKind(KindInfo, MsgBuilt)
	s.saveWorld(p)
}

// cmdName: "@name The Dark Cave" – the room's name.
func cmdName(s *Server, p *Player, cmd Command) {
	if cmd.Rest == "" {
		p.SendKind(KindError, MsgNameUsage)
		return
	}
	p.Room.Name = cmd.Rest
	p.SendKind(KindInfo, MsgBuilt)
	s.saveWorld(p)
}

// cmdExit: "@exit east shop" adds an exit (one way), "@exit east none" removes it.
func cmdExit(s *Server, p *Player, cmd Command) {
	// TODO(lesson 36): "<direction> <room id>" or "<direction> none" (MsgExitUsage),
	// one way, through s.change; rezone; MsgBuilt; saveWorld.
}

// worldSnapshot is the world as a world file. Runs in the game loop.
func (s *Server) worldSnapshot() worldFile {
	w := s.world
	f := worldFile{Start: w.Start, ZoneSize: w.ZoneSize, Seed: w.Seed}
	for _, r := range w.Rooms {
		f.Rooms = append(f.Rooms, r)
	}
	sort.Slice(f.Rooms, func(i, j int) bool { return f.Rooms[i].ID < f.Rooms[j].ID })
	for _, it := range w.Items {
		f.Items = append(f.Items, it)
	}
	sort.Slice(f.Items, func(i, j int) bool { return f.Items[i].ID < f.Items[j].ID })
	for _, n := range w.NPCs {
		f.NPCs = append(f.NPCs, n)
	}
	sort.Slice(f.NPCs, func(i, j int) bool { return f.NPCs[i].ID < f.NPCs[j].ID })
	return f
}

// saveWorld writes the world file (if there is one). The JSON is made in the
// game loop, the disk is written outside it. The file is written to a
// temporary file first and then renamed, so nobody – not even the watcher
// from lesson 31 – ever reads half a file.
func (s *Server) saveWorld(p *Player) {
	// TODO(lesson 36): MsgNotSaved without a world file. Otherwise the JSON of
	// worldSnapshot here, in the loop; then in a goroutine, under s.saveMu: skip
	// it if a newer snapshot (saveSeq/savedSeq) is already written,
	// writeFileAtomic, and remember the file's stamp in s.ownWrite.
	_, _, _ = json.MarshalIndent, log.Printf, os.Stat
}

// writeFileAtomic writes data to a temporary file next to path and renames it.
func writeFileAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), path)
}
