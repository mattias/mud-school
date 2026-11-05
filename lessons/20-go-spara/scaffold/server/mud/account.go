package mud

import (
	"log"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

// MinPasswordLen is the shortest password a new account may have (in characters).
const MinPasswordLen = 4

// PasswordCost is how much work bcrypt does per password: each step up doubles
// the time. Slow is good – it makes guessing passwords expensive.
var PasswordCost = bcrypt.DefaultCost

// Prompt kinds for {"type":"prompt"} frames, so that a graphical client can
// hide what is typed. See PROTOCOL.md.
const (
	PromptPassword    = "password"
	PromptNewPassword = "new_password"
)

// askPassword sends a question and tells a graphical client that the answer is secret.
func (p *Player) askPassword(text, prompt string) {
	p.Send(text)
	p.SendFrame(Frame{Type: "prompt", Prompt: prompt})
}

// authenticate asks for the password of an existing account, or lets a new
// player choose one. It runs in the player's own goroutine – bcrypt is slow on
// purpose, and the game loop must not wait for it.
// ok is false if the player disconnected; retry is true if the name should be asked again.
func (s *Server) authenticate(p *Player) (ok, retry bool) {
	// TODO:
	//   - s.store.PasswordHash(p.Name); ErrNoAccount -> return s.register(p)
	//   - p.askPassword(MsgAskPassword, PromptPassword), read the answer
	//   - bcrypt.CompareHashAndPassword; wrong -> MsgWrongPassword, retry
	return false, false
}

// register lets a new player choose a password and creates the account.
func (s *Server) register(p *Player) (ok, retry bool) {
	// TODO: loop: ask with MsgChoosePassword / PromptNewPassword, at least
	// MinPasswordLen characters (runes!), bcrypt.GenerateFromPassword with
	// PasswordCost, s.store.CreateAccount.
	_ = utf8.RuneCountInString
	_ = log.Printf
	return false, false
}

// snapshot is what will be saved about p. Runs in the game loop.
func (p *Player) snapshot() SavedPlayer {
	sp := SavedPlayer{Name: p.Name, HP: p.HP}
	if p.Room != nil {
		sp.Room = p.Room.ID
	}
	for _, it := range p.Inventory {
		sp.Items = append(sp.Items, it.ID)
	}
	return sp
}

// restore puts a saved player back into the world. Runs in the game loop.
// Items are taken from wherever they lie; an item someone else is carrying stays with them.
func (s *Server) restore(p *Player, sp SavedPlayer) {
	// TODO: the room (if it still exists), the HP (if it is sensible) and
	// the items (with s.takeItem).
}

// takeItem removes the item with the given ID from the room it lies in.
// nil if no room has it.
func (s *Server) takeItem(id string) *Item {
	it := s.world.Items[id]
	if it == nil {
		return nil
	}
	for _, r := range s.world.Rooms {
		for _, x := range r.Items {
			if x == it {
				r.Items = remove(r.Items, it)
				return it
			}
		}
	}
	return nil
}
