package mud

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

// NPC is a creature in the world: it has no connection, but it can wander,
// talk and (from lesson 18) fight.
type NPC struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"` // e.g. "a black cat"
	Description string         `json:"description"`
	Keywords    []string       `json:"keywords"`
	Room        string         `json:"room"`            // where it starts
	Behaviors   []BehaviorSpec `json:"behaviors"`       // what it does every tick
	Glyph       string         `json:"glyph,omitempty"` // how a text client draws it (lesson 23); default: the first letter of the ID

	// For combat in lesson 18 (unused until then).
	HP         int    `json:"hp,omitempty"`
	Damage     [2]int `json:"damage,omitempty"` // min, max
	Attackable bool   `json:"attackable,omitempty"`

	// Owned by the game loop.
	Location  *Room      `json:"-"`
	Pos       Pos        `json:"-"` // where it stands in the zone
	behaviors []Behavior // built from Behaviors
	origin    *Room      // where it comes back after a defeat
	home      Pos        // where in origin's zone it comes back
	hp        int        // current health
	dead      bool       // defeated, waiting to respawn
	respawnAt int64      // the tick when it comes back
}

// Matches reports whether the player's word refers to the creature.
func (n *NPC) Matches(word string) bool {
	word = strings.ToLower(word)
	if strings.ToLower(n.ID) == word {
		return true
	}
	for _, k := range n.Keywords {
		if strings.ToLower(k) == word {
			return true
		}
	}
	return false
}

// AddBehavior gives the creature one more behaviour.
func (n *NPC) AddBehavior(b Behavior) {
	n.behaviors = append(n.behaviors, b)
}

// Behavior is something a creature does. Act runs once per tick, in the
// game loop.
type Behavior interface {
	Act(s *Server, n *NPC)
}

// BehaviorSpec is how a behaviour is written in npcs.json:
//
//	{"type": "wander", "chance": 0.1}
//	{"type": "talk", "chance": 0.05, "lines": ["Meow."]}
type BehaviorSpec struct {
	Type   string   `json:"type"`
	Chance float64  `json:"chance"`          // probability per tick, 0–1
	Lines  []string `json:"lines,omitempty"` // for "talk"
	Range  int      `json:"range,omitempty"` // for "hunt": how far it sees (lesson 27)
	Below  float64  `json:"below,omitempty"` // for "flee": flee below this share of its health
}

func newBehavior(spec BehaviorSpec) (Behavior, bool) {
	switch spec.Type {
	case "wander":
		return Wanderer{Chance: spec.Chance}, true
	case "talk":
		return Talker{Chance: spec.Chance, Lines: spec.Lines}, true
	case "hunt":
		return Hunter{Range: spec.Range, Chance: spec.Chance}, true
	case "flee":
		return Fleer{Below: spec.Below}, true
	}
	return nil, false
}

// Wanderer walks through a random exit now and then.
type Wanderer struct{ Chance float64 }

func (b Wanderer) Act(s *Server, n *NPC) {
	if s.fighting(n) {
		return // creatures don't wander off in the middle of a fight
	}
	if s.rng.Float64() >= b.Chance {
		return
	}
	var dirs []string
	for _, dir := range Directions { // fixed order: the same seed gives the same walk
		if _, ok := n.Location.Exits[dir]; ok {
			dirs = append(dirs, dir)
		}
	}
	if len(dirs) == 0 {
		return
	}
	dir := dirs[s.rng.IntN(len(dirs))]
	to := s.world.Room(n.Location.Exits[dir])
	if to == nil {
		return
	}
	s.moveNPC(n, to, dir)
}

// Talker says one of its lines now and then.
type Talker struct {
	Chance float64
	Lines  []string
}

func (b Talker) Act(s *Server, n *NPC) {
	if len(b.Lines) == 0 || s.rng.Float64() >= b.Chance {
		return
	}
	line := b.Lines[s.rng.IntN(len(b.Lines))]
	s.broadcastRoom(n.Location, nil, KindSay, fmt.Sprintf(MsgSays, n.Name, line))
}

// moveNPC moves a creature and tells both rooms. Runs in the game loop.
func (s *Server) moveNPC(n *NPC, to *Room, dir string) {
	from := n.Location
	s.broadcastRoom(from, nil, KindMove, fmt.Sprintf(MsgLeaves, n.Name, directionTravel[dir]))
	from.NPCs = remove(from.NPCs, n)
	n.Location = to
	n.Pos = s.world.zoneOf(to).arrival(opposite[dir])
	to.NPCs = append(to.NPCs, n)
	s.broadcastRoom(to, nil, KindMove, fmt.Sprintf(MsgArrives, n.Name))
}

// npcsFile is the format of npcs.json (and of "npcs" in a world file).
type npcsFile struct {
	NPCs []*NPC `json:"npcs"`
}

// LoadNPCs reads creatures in JSON format and puts them in their rooms.
func (w *World) LoadNPCs(r io.Reader) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var f npcsFile
	if err := dec.Decode(&f); err != nil {
		return fmt.Errorf(MsgErrReadWorld, err)
	}
	return w.AddNPCs(f.NPCs...)
}

// AddNPCs validates creatures, builds their behaviours and puts them in
// their rooms. It reports all problems at once.
func (w *World) AddNPCs(npcs ...*NPC) error {
	if w.NPCs == nil {
		w.NPCs = make(map[string]*NPC)
	}
	var errs []error
	for _, n := range npcs {
		switch {
		case n.ID == "" || n.Name == "":
			errs = append(errs, fmt.Errorf(MsgErrNPCFields, n.ID))
			continue
		case w.NPCs[n.ID] != nil:
			errs = append(errs, fmt.Errorf(MsgErrDuplicateNPC, n.ID))
			continue
		case w.Room(n.Room) == nil:
			errs = append(errs, fmt.Errorf(MsgErrNPCRoom, n.ID, n.Room))
			continue
		}
		ok := true
		for _, spec := range n.Behaviors {
			b, known := newBehavior(spec)
			if !known {
				errs = append(errs, fmt.Errorf(MsgErrBehavior, n.ID, spec.Type))
				ok = false
				continue
			}
			n.AddBehavior(b)
		}
		if !ok {
			continue
		}
		if n.Attackable && n.HP <= 0 {
			n.HP = 1
		}
		n.hp = n.HP
		w.NPCs[n.ID] = n
		n.Location = w.Room(n.Room)
		n.origin = n.Location
		n.Pos = w.zoneOf(n.Location).freeNear(w.zoneOf(n.Location).Center(), w.taken(n.Location))
		n.home = n.Pos
		n.Location.NPCs = append(n.Location.NPCs, n)
	}
	return errors.Join(errs...)
}

// npcsByID returns the world's creatures in a fixed order, so that a seeded
// random generator gives the same result every time.
func (w *World) npcsByID() []*NPC {
	list := make([]*NPC, 0, len(w.NPCs))
	for _, n := range w.NPCs {
		list = append(list, n)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	return list
}
