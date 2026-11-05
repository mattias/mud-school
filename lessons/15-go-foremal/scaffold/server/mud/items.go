package mud

import (
	"encoding/json"
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
	// TODO(lesson 15): true if word equals the ID or one of the keywords,
	// ignoring upper/lower case.
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
	// TODO(lesson 15)
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
	// TODO(lesson 15): check every item (MsgErrItemFields, MsgErrDuplicateItem,
	// MsgErrItemRoom), put the good ones in w.Items and in their room's Items,
	// and return all problems with errors.Join.
	return nil
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
	// TODO(lesson 15), see the README
}

func cmdDrop(s *Server, p *Player, cmd Command) {
	// TODO(lesson 15), see the README
}

func cmdInventory(s *Server, p *Player, cmd Command) {
	// TODO(lesson 15), see the README
}

func cmdExamine(s *Server, p *Player, cmd Command) {
	// TODO(lesson 15), see the README
}
