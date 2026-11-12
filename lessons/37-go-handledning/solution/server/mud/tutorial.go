package mud

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
)

// The tutorial (lesson 37): new players start in a small area of their own
// and are guided, step by step, through the first commands. The steps are
// data in the world file; the server only checks, after every command,
// whether the player has done what the current step asks for.

// Tutorial is the "tutorial" object of a world file.
type Tutorial struct {
	Start string         `json:"start"` // the room new players start in
	Steps []TutorialStep `json:"steps"`
	Done  string         `json:"done"` // said when the last step is done
}

// TutorialStep is one step. Exactly one of Command, Room and Item says when
// it is done.
type TutorialStep struct {
	Text    string `json:"text"`
	Command string `json:"command,omitempty"` // done when the player runs this verb, e.g. "look"
	Room    string `json:"room,omitempty"`    // done when the player is in this room
	Item    string `json:"item,omitempty"`    // done when the player carries this item
}

// tutorialFile is the format of tutorial.json: rooms, items and the steps.
type tutorialFile struct {
	Rooms    []*Room   `json:"rooms"`
	Items    []*Item   `json:"items,omitempty"`
	Tutorial *Tutorial `json:"tutorial"`
}

// LoadTutorial adds the tutorial area from a tutorial file to the world.
func (w *World) LoadTutorial(r io.Reader) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var f tutorialFile
	if err := dec.Decode(&f); err != nil {
		return fmt.Errorf(MsgErrReadWorld, err)
	}
	for _, room := range f.Rooms {
		if w.Room(room.ID) != nil {
			return fmt.Errorf(MsgErrDuplicateRoom, room.ID)
		}
		w.Rooms[room.ID] = room
	}
	if err := w.Validate(); err != nil {
		return err
	}
	if err := w.AddItems(f.Items...); err != nil {
		return err
	}
	w.Tutorial = f.Tutorial
	return w.validateTutorial()
}

// tutorialArea is every room you can walk to from the tutorial's start.
func (w *World) tutorialArea() map[string]bool {
	area := map[string]bool{}
	if w.Tutorial == nil || w.Room(w.Tutorial.Start) == nil {
		return area
	}
	queue := []string{w.Tutorial.Start}
	area[w.Tutorial.Start] = true
	for len(queue) > 0 {
		r := w.Room(queue[0])
		queue = queue[1:]
		for _, to := range r.Exits {
			if !area[to] && w.Room(to) != nil {
				area[to] = true
				queue = append(queue, to)
			}
		}
	}
	return area
}

// validateTutorial checks the tutorial against the world and returns all
// the problems (nil if there is no tutorial). Items must already be added.
func (w *World) validateTutorial() error {
	t := w.Tutorial
	if t == nil {
		return nil
	}
	var errs []error
	if w.Room(t.Start) == nil {
		return fmt.Errorf(MsgErrTutorialStart, t.Start)
	}
	area := w.tutorialArea()
	if area[w.Start] {
		errs = append(errs, fmt.Errorf(MsgErrTutorialOpen, w.Start))
	}
	if len(t.Steps) == 0 {
		errs = append(errs, errors.New(MsgErrTutorialEmpty))
	}
	for i, st := range t.Steps {
		n := 0
		for _, set := range []bool{st.Command != "", st.Room != "", st.Item != ""} {
			if set {
				n++
			}
		}
		switch {
		case st.Text == "" || n != 1:
			errs = append(errs, fmt.Errorf(MsgErrTutorialStep, i+1))
		case st.Room != "" && !area[st.Room]:
			errs = append(errs, fmt.Errorf(MsgErrTutorialRoom, i+1, st.Room))
		case st.Item != "" && (w.Items[st.Item] == nil || !area[w.Items[st.Item].Room]):
			errs = append(errs, fmt.Errorf(MsgErrTutorialItem, i+1, st.Item))
		}
	}
	return errors.Join(errs...)
}

// startTutorial puts a new player at the start of the tutorial (if the world
// has one). Runs in the game loop.
func (s *Server) startTutorial(p *Player) {
	t := s.world.Tutorial
	if t == nil {
		return
	}
	r := s.world.Room(t.Start)
	p.Room = r
	p.Pos = s.world.zoneOf(r).freeNear(s.world.zoneOf(r).Center(), nil)
	p.tutorial = 1
}

// showTutorialStep tells the player what to do now.
func (s *Server) showTutorialStep(p *Player) {
	t := s.world.Tutorial
	if p.tutorial == 0 || t == nil || p.tutorial > len(t.Steps) {
		return
	}
	p.SendKind(KindInfo, fmt.Sprintf(MsgTutorialStep, p.tutorial, len(t.Steps), t.Steps[p.tutorial-1].Text))
}

// stepDone reports whether p has done what step st asks for. verb is the
// command that was just run.
func (s *Server) stepDone(p *Player, st TutorialStep, verb string) bool {
	switch {
	case st.Command != "":
		return verb == st.Command
	case st.Room != "":
		return p.Room.ID == st.Room
	case st.Item != "":
		return slices.ContainsFunc(p.Inventory, func(it *Item) bool { return it.ID == st.Item })
	}
	return false
}

// onTutorialCommand is a hook (lesson 34) for EventCommand: after every
// command it checks the player's current step – and the next, since one
// command can finish several.
func (s *Server) onTutorialCommand(e *Event) {
	p := e.Player
	if p == nil || p.tutorial == 0 {
		return
	}
	t := s.world.Tutorial
	if t == nil { // the world was reloaded without a tutorial: out to the start
		p.tutorial = 0
		s.moveToStart(p)
		return
	}
	advanced := false
	for p.tutorial > 0 && p.tutorial <= len(t.Steps) && s.stepDone(p, t.Steps[p.tutorial-1], e.Text) {
		p.tutorial++
		advanced = true
	}
	if p.tutorial > len(t.Steps) {
		p.SendKind(KindInfo, t.Done)
		s.leaveTutorial(p)
		return
	}
	if advanced {
		s.showTutorialStep(p)
	}
}

// leaveTutorial takes the player out of the tutorial, to the world's start
// room. Things from the tutorial area go back where they were, so that the
// next new player finds them too.
func (s *Server) leaveTutorial(p *Player) {
	area := s.world.tutorialArea()
	p.tutorial = 0
	kept := p.Inventory[:0]
	for _, it := range p.Inventory {
		if area[it.Room] {
			s.world.placeItem(s.world.Room(it.Room), it)
		} else {
			kept = append(kept, it)
		}
	}
	p.Inventory = kept
	if area[p.Room.ID] { // not if they are already out (e.g. they skipped after dying)
		s.moveToStart(p)
	}
}

// moveToStart moves the player to the world's start room, in front of everyone.
func (s *Server) moveToStart(p *Player) {
	start := s.world.Room(s.world.Start)
	s.broadcastRoom(p.Room, p, KindMove, fmt.Sprintf(MsgDisappears, p.Name))
	p.Room = start
	p.Pos = s.world.zoneOf(start).freeNear(s.world.zoneOf(start).Center(), nil)
	s.broadcastRoom(start, p, KindMove, fmt.Sprintf(MsgArrives, p.Name))
	s.look(p)
}

// cmdTutorial: "tutorial" repeats the current step, "tutorial skip" leaves.
func cmdTutorial(s *Server, p *Player, cmd Command) {
	if p.tutorial == 0 {
		p.SendKind(KindError, MsgTutorialNotIn)
		return
	}
	if cmd.Rest == "skip" {
		p.SendKind(KindInfo, MsgTutorialSkipped)
		s.leaveTutorial(p)
		return
	}
	s.showTutorialStep(p)
}
