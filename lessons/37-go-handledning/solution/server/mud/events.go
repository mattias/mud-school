package mud

import (
	"log"

	lua "github.com/yuin/gopher-lua"
)

// Events and hooks (lesson 34): the server announces what is about to
// happen – "Anna is entering the vault" – and hooks may react, or say no.
// Go code, Lua scripts and (as a challenge) plugins can all subscribe.

// The kinds of events.
const (
	EventEnter = "enter" // a player is about to walk into another room (can be cancelled)
	EventSay   = "say"   // a player is about to say something (can be cancelled, Text can be changed)
	EventTick  = "tick"  // a tick has passed
	EventDeath = "death" // a player or a creature has died
	// EventCommand: a player has run a command (lesson 37). Text is the verb.
	// It comes afterwards, so it can't be cancelled.
	EventCommand = "command"
)

// Event is something that happens in the game.
type Event struct {
	Kind   string
	Player *Player // who does it (nil for a tick, or when a creature dies of something else)
	NPC    *NPC    // a creature involved (death)
	Room   *Room   // where it happens
	To     *Room   // enter: the room the player is walking into
	Dir    string  // enter: the direction
	Text   string  // say: what is said; hooks may change it

	cancelled bool
	reason    string
}

// Cancel stops the action. reason is told to the player.
func (e *Event) Cancel(reason string) {
	e.cancelled = true
	e.reason = reason
}

// Cancelled reports whether a hook has cancelled the event.
func (e *Event) Cancelled() bool { return e.cancelled }

// Hook reacts to an event.
type Hook func(e *Event)

// On subscribes a hook to a kind of event. Call it before the server starts
// or from inside the game loop.
func (s *Server) On(kind string, h Hook) {
	if s.hooks == nil {
		s.hooks = map[string][]Hook{}
	}
	s.hooks[kind] = append(s.hooks[kind], h)
}

// emit runs the hooks for an event – Go hooks first, then Lua hooks – and
// stops at the first that cancels. It returns false if the event was
// cancelled. Runs in the game loop.
func (s *Server) emit(e *Event) bool {
	for _, h := range s.hooks[e.Kind] {
		h(e)
		if e.cancelled {
			return false
		}
	}
	if s.scripts != nil {
		for _, fn := range s.scripts.hooks[e.Kind] {
			reason, err := s.scripts.runHook(fn, e)
			if err != nil {
				log.Printf("script hook %s: %v", e.Kind, err)
				continue
			}
			if reason != "" {
				e.Cancel(reason)
				return false
			}
		}
	}
	return true
}

// refuse tells the player why an action was cancelled.
func (e *Event) refuse() {
	if e.Player == nil {
		return
	}
	reason := e.reason
	if reason == "" {
		reason = MsgCancelled
	}
	e.Player.SendKind(KindError, reason)
}

// runHook calls a Lua hook with the event as a table:
//
//	{kind=, player=, npc=, room=, to=, dir=, text=}
//
// A hook that returns a string cancels the event with that reason.
func (sc *Scripts) runHook(fn *lua.LFunction, e *Event) (string, error) {
	t := sc.L.NewTable()
	t.RawSetString("kind", lua.LString(e.Kind))
	if e.Player != nil {
		t.RawSetString("player", lua.LString(e.Player.Name))
	}
	if e.NPC != nil {
		t.RawSetString("npc", lua.LString(e.NPC.ID))
	}
	if e.Room != nil {
		t.RawSetString("room", lua.LString(e.Room.ID))
	}
	if e.To != nil {
		t.RawSetString("to", lua.LString(e.To.ID))
	}
	t.RawSetString("dir", lua.LString(e.Dir))
	t.RawSetString("text", lua.LString(e.Text))

	ret, err := sc.call(fn, 1, t)
	if err != nil {
		return "", err
	}
	if s, ok := ret.(lua.LString); ok {
		return string(s), nil
	}
	return "", nil
}
