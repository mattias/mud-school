package mud

import (
	"errors"
	"fmt"
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
	hash, err := s.store.PasswordHash(p.Name)
	switch {
	case errors.Is(err, ErrNoAccount):
		return s.register(p)
	case err != nil:
		log.Printf("cannot read the account %s: %v", p.Name, err)
		p.SendKind(KindError, MsgStoreError)
		return false, false
	}
	p.askPassword(MsgAskPassword, PromptPassword)
	pw, ok := p.readLine()
	if !ok {
		return false, false
	}
	if bcrypt.CompareHashAndPassword(hash, []byte(pw)) != nil {
		p.SendKind(KindError, MsgWrongPassword)
		return true, true
	}
	return true, false
}

// register lets a new player choose a password and creates the account.
func (s *Server) register(p *Player) (ok, retry bool) {
	for {
		p.askPassword(fmt.Sprintf(MsgChoosePassword, p.Name), PromptNewPassword)
		pw, ok := p.readLine()
		if !ok {
			return false, false
		}
		if utf8.RuneCountInString(pw) < MinPasswordLen {
			p.SendKind(KindError, MsgPasswordTooShort)
			continue
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(pw), PasswordCost)
		if err != nil {
			log.Printf("bcrypt: %v", err)
			return false, false
		}
		if err := s.store.CreateAccount(p.Name, hash); errors.Is(err, ErrNameTaken) {
			// Someone else registered the name a moment ago.
			p.Sendf(MsgInvalidName, err)
			return true, true
		} else if err != nil {
			log.Printf("cannot create the account %s: %v", p.Name, err)
			p.SendKind(KindError, MsgStoreError)
			return false, false
		}
		return true, false
	}
}

// snapshot is what will be saved about p. Runs in the game loop.
func (p *Player) snapshot() SavedPlayer {
	sp := SavedPlayer{Name: p.Name, HP: p.HP, Tutorial: p.tutorial, Language: p.lang}
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
	if r := s.world.Room(sp.Room); r != nil {
		p.Room = r
		p.Pos = s.world.zoneOf(r).freeNear(s.world.zoneOf(r).Center(), nil)
	}
	if t := s.world.Tutorial; t != nil && sp.Tutorial > 0 {
		p.tutorial = min(sp.Tutorial, len(t.Steps)) // lesson 37; a shorter tutorial: its last step
	}
	if p.tutorial == 0 && s.world.tutorialArea()[p.Room.ID] {
		// In the tutorial's rooms but not in the tutorial: they could never get out.
		start := s.world.Room(s.world.Start)
		p.Room = start
		p.Pos = s.world.zoneOf(start).freeNear(s.world.zoneOf(start).Center(), nil)
	}
	p.lang = sp.Language // lesson 38
	if sp.HP > 0 && sp.HP <= p.MaxHP {
		p.HP = sp.HP
	}
	for _, id := range sp.Items {
		if it := s.takeItem(id); it != nil {
			p.Inventory = append(p.Inventory, it)
		}
	}
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
