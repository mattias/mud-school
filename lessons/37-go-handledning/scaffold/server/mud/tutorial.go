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
	// TODO(lesson 37): decode a tutorialFile (DisallowUnknownFields, errors as
	// MsgErrReadWorld), add its rooms to w.Rooms (an ID that is already taken:
	// MsgErrDuplicateRoom), w.Validate(), w.AddItems(...), set w.Tutorial and
	// finish with w.validateTutorial().
	_ = json.NewDecoder
	return errors.New("TODO: LoadTutorial")
}

// tutorialArea is every room you can walk to from the tutorial's start.
func (w *World) tutorialArea() map[string]bool {
	// TODO(lesson 37): breadth-first search over the exits from
	// w.Tutorial.Start (if there is a tutorial and the room exists).
	return map[string]bool{}
}

// validateTutorial checks the tutorial against the world and returns all
// the problems (nil if there is no tutorial). Items must already be added.
func (w *World) validateTutorial() error {
	// TODO(lesson 37): nil without a tutorial. MsgErrTutorialStart if the start
	// room is missing (stop there). Then collect: MsgErrTutorialOpen if the
	// world's start is in the tutorial area, MsgErrTutorialEmpty without steps,
	// and per step (numbered from 1) MsgErrTutorialStep (no text, or not exactly
	// one of command/room/item), MsgErrTutorialRoom, MsgErrTutorialItem.
	return nil
}

// startTutorial puts a new player at the start of the tutorial (if the world
// has one). Runs in the game loop.
func (s *Server) startTutorial(p *Player) {
	// TODO(lesson 37): with a tutorial: the start room, a free cell near the
	// middle of its zone, and p.tutorial = 1.
}

// showTutorialStep tells the player what to do now.
func (s *Server) showTutorialStep(p *Player) {
	// TODO(lesson 37): MsgTutorialStep with the step's number, the number of
	// steps and its text – if the player is in the tutorial.
	_ = fmt.Sprintf
}

// stepDone reports whether p has done what step st asks for. verb is the
// command that was just run.
func (s *Server) stepDone(p *Player, st TutorialStep, verb string) bool {
	// TODO(lesson 37): command: verb == st.Command; room: the player is in it;
	// item: the player carries it (slices.ContainsFunc).
	_ = slices.ContainsFunc[[]*Item]
	return false
}

// onTutorialCommand is a hook (lesson 34) for EventCommand: after every
// command it checks the player's current step – and the next, since one
// command can finish several.
func (s *Server) onTutorialCommand(e *Event) {
	// TODO(lesson 37): for a player in the tutorial: advance while the current
	// step is done (a world without the tutorial, or with fewer steps: just
	// p.tutorial = 0). After the last step: t.Done and leaveTutorial; otherwise,
	// if something was done, show the new step.
}

// leaveTutorial takes the player out of the tutorial, to the world's start
// room. Things from the tutorial area go back where they were, so that the
// next new player finds them too.
func (s *Server) leaveTutorial(p *Player) {
	// TODO(lesson 37): p.tutorial = 0; things from the tutorial area go back to
	// their rooms (placeItem); if the player is in the area: MsgDisappears to
	// the room, to the world's start room (a free cell), MsgArrives, look.
}

// cmdTutorial: "tutorial" repeats the current step, "tutorial skip" leaves.
func cmdTutorial(s *Server, p *Player, cmd Command) {
	// TODO(lesson 37): MsgTutorialNotIn outside the tutorial; "tutorial skip":
	// MsgTutorialSkipped and leaveTutorial; otherwise showTutorialStep.
}
