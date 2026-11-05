package mud

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Item is a thing that lies in a room or is carried by a player.
type Item struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`        // e.g. "a rusty sword"
	Description string   `json:"description"` // shown by examine
	Keywords    []string `json:"keywords"`    // words the player can use: "sword", "rusty"
	Room        string   `json:"room"`        // where it lies when the world starts
}

// Matches reports whether the player's word refers to the item.
func (i *Item) Matches(word string) bool {
	word = strings.ToLower(word)
	if strings.ToLower(i.ID) == word {
		return true
	}
	for _, k := range i.Keywords {
		if strings.ToLower(k) == word {
			return true
		}
	}
	return false
}

// Matcher is anything a player can refer to with a word: items now, creatures
// in lesson 16.
type Matcher interface {
	Matches(word string) bool
}

// FindMatch returns the first element of list that matches word.
// It is generic: it works for []*Item, []*NPC, ... – any type that has a
// Matches method.
func FindMatch[T Matcher](list []T, word string) (T, bool) {
	for _, x := range list {
		if x.Matches(word) {
			return x, true
		}
	}
	var zero T
	return zero, false
}

// remove returns list without x.
func remove[T comparable](list []T, x T) []T {
	for i, y := range list {
		if y == x {
			return append(list[:i:i], list[i+1:]...)
		}
	}
	return list
}

// itemsFile is the format of items.json (and of "items" in a world file).
type itemsFile struct {
	Items []*Item `json:"items"`
}

// LoadItems reads items in JSON format and places them in the world's rooms.
func (w *World) LoadItems(r io.Reader) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var f itemsFile
	if err := dec.Decode(&f); err != nil {
		return fmt.Errorf(MsgErrReadWorld, err)
	}
	return w.AddItems(f.Items...)
}

// AddItems validates items and places them in their rooms. It reports all
// problems at once, like Validate.
func (w *World) AddItems(items ...*Item) error {
	if w.Items == nil {
		w.Items = make(map[string]*Item)
	}
	var errs []error
	for _, it := range items {
		switch {
		case it.ID == "" || it.Name == "":
			errs = append(errs, fmt.Errorf(MsgErrItemFields, it.ID))
		case w.Items[it.ID] != nil:
			errs = append(errs, fmt.Errorf(MsgErrDuplicateItem, it.ID))
		case w.Room(it.Room) == nil:
			errs = append(errs, fmt.Errorf(MsgErrItemRoom, it.ID, it.Room))
		default:
			w.Items[it.ID] = it
			room := w.Room(it.Room)
			room.Items = append(room.Items, it)
		}
	}
	return errors.Join(errs...)
}

// itemNames joins the items' names for a message.
func itemNames(items []*Item) string {
	names := make([]string, len(items))
	for i, it := range items {
		names[i] = it.Name
	}
	return strings.Join(names, MsgListSep)
}

// The item commands run in the game loop.

func cmdGet(s *Server, p *Player, cmd Command) {
	if cmd.Rest == "" {
		p.SendKind(KindError, MsgGetWhat)
		return
	}
	it, ok := FindMatch(p.Room.Items, cmd.Rest)
	if !ok {
		p.SendKind(KindError, fmt.Sprintf(MsgNoSuchItem, cmd.Rest))
		return
	}
	p.Room.Items = remove(p.Room.Items, it)
	p.Inventory = append(p.Inventory, it)
	p.SendKind(KindInfo, fmt.Sprintf(MsgYouGet, it.Name))
	s.broadcastRoom(p.Room, p, KindInfo, fmt.Sprintf(MsgGets, p.Name, it.Name))
}

func cmdDrop(s *Server, p *Player, cmd Command) {
	if cmd.Rest == "" {
		p.SendKind(KindError, MsgDropWhat)
		return
	}
	it, ok := FindMatch(p.Inventory, cmd.Rest)
	if !ok {
		p.SendKind(KindError, fmt.Sprintf(MsgNotCarrying, cmd.Rest))
		return
	}
	p.Inventory = remove(p.Inventory, it)
	p.Room.Items = append(p.Room.Items, it)
	p.SendKind(KindInfo, fmt.Sprintf(MsgYouDrop, it.Name))
	s.broadcastRoom(p.Room, p, KindInfo, fmt.Sprintf(MsgDrops, p.Name, it.Name))
}

func cmdInventory(s *Server, p *Player, cmd Command) {
	if len(p.Inventory) == 0 {
		p.SendKind(KindInfo, MsgInventoryEmpty)
		return
	}
	p.SendKind(KindInfo, fmt.Sprintf(MsgInventory, itemNames(p.Inventory)))
}

func cmdExamine(s *Server, p *Player, cmd Command) {
	if cmd.Rest == "" {
		p.SendKind(KindError, MsgExamineWhat)
		return
	}
	if it, ok := FindMatch(p.Inventory, cmd.Rest); ok {
		p.SendKind(KindInfo, it.Description)
		return
	}
	if it, ok := FindMatch(p.Room.Items, cmd.Rest); ok {
		p.SendKind(KindInfo, it.Description)
		return
	}
	p.SendKind(KindError, fmt.Sprintf(MsgNoSuchItem, cmd.Rest))
}
