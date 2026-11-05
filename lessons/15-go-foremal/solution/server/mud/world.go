package mud

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
)

// Room is a place in the world. The `json:"..."` tags control what the
// fields are called in the JSON file.
type Room struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Terrain     string            `json:"terrain"`         // one of Terrains
	X           int               `json:"x"`               // east is +1
	Y           int               `json:"y"`               // south is +1
	Z           int               `json:"z"`               // up is +1
	Exits       map[string]string `json:"exits,omitempty"` // direction -> room ID

	Items []*Item `json:"-"` // what lies here right now (owned by the game loop)
}

// World is all the rooms in the game.
type World struct {
	Start string           // ID of the room where new players appear
	Rooms map[string]*Room // ID -> room
	Items map[string]*Item // ID -> item, wherever it is now
}

// worldFile is the format of a world file.
type worldFile struct {
	Start string  `json:"start"`
	Rooms []*Room `json:"rooms"`
	Items []*Item `json:"items,omitempty"` // optional; the default world keeps them in items.json
}

// Directions are all the directions in the order they are shown.
// (Their names as shown to players are in messages.go.)
var Directions = []string{"north", "east", "south", "west", "up", "down"}

// Terrains are the kinds of terrain a room can have.
var Terrains = []string{"town", "inside", "road", "field", "forest", "water", "mountain", "cave"}

// directionDelta is how the coordinates change when you walk in a direction.
var directionDelta = map[string][3]int{
	"north": {0, -1, 0}, "south": {0, 1, 0}, "east": {1, 0, 0}, "west": {-1, 0, 0}, "up": {0, 0, 1}, "down": {0, 0, -1},
}

// NewWorld creates a world from the rooms. "..." makes the function variadic:
// NewWorld("square", room1, room2, room3) – rooms becomes a []*Room.
func NewWorld(start string, rooms ...*Room) *World {
	w := &World{Start: start, Rooms: make(map[string]*Room, len(rooms))}
	for _, r := range rooms {
		w.Rooms[r.ID] = r
	}
	return w
}

// Room returns the room with the given ID, or nil if there is none.
func (w *World) Room(id string) *Room {
	return w.Rooms[id]
}

// LoadWorld reads and validates a world in JSON format.
func LoadWorld(r io.Reader) (*World, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields() // typos like "nmae" must be errors, not silently ignored
	var f worldFile
	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf(MsgErrReadWorld, err)
	}
	seen := make(map[string]bool)
	for _, room := range f.Rooms {
		if seen[room.ID] {
			return nil, fmt.Errorf(MsgErrDuplicateRoom, room.ID)
		}
		seen[room.ID] = true
	}
	w := NewWorld(f.Start, f.Rooms...)
	if err := w.Validate(); err != nil {
		return nil, err
	}
	if err := w.AddItems(f.Items...); err != nil {
		return nil, err
	}
	return w, nil
}

// LoadWorldFile reads a world from a file.
func LoadWorldFile(path string) (*World, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadWorld(f)
}

// Validate looks for mistakes in the world and returns ALL of them, joined
// with errors.Join (or nil if everything is fine).
func (w *World) Validate() error {
	var errs []error
	if w.Room(w.Start) == nil {
		errs = append(errs, fmt.Errorf(MsgErrStartMissing, w.Start))
	}

	ids := make([]string, 0, len(w.Rooms))
	for id := range w.Rooms {
		ids = append(ids, id)
	}
	sort.Strings(ids) // stable order of the errors

	positions := make(map[[3]int]string)
	for _, id := range ids {
		r := w.Rooms[id]
		if r.ID == "" || r.Name == "" {
			errs = append(errs, fmt.Errorf(MsgErrRoomFields, id))
		}
		if !slices.Contains(Terrains, r.Terrain) {
			errs = append(errs, fmt.Errorf(MsgErrTerrain, id, r.Terrain))
		}
		pos := [3]int{r.X, r.Y, r.Z}
		if other, taken := positions[pos]; taken {
			errs = append(errs, fmt.Errorf(MsgErrSamePos, id, pos, other))
		}
		positions[pos] = id

		for _, dir := range Directions {
			to, ok := r.Exits[dir]
			if !ok {
				continue
			}
			target := w.Room(to)
			if target == nil {
				errs = append(errs, fmt.Errorf(MsgErrExitMissing, id, dir, to))
				continue
			}
			d := directionDelta[dir]
			if target.X != r.X+d[0] || target.Y != r.Y+d[1] || target.Z != r.Z+d[2] {
				errs = append(errs, fmt.Errorf(MsgErrExitCoords, id, dir, to))
			}
		}
		for dir := range r.Exits {
			if _, ok := directionDelta[dir]; !ok {
				errs = append(errs, fmt.Errorf(MsgErrDirection, id, dir))
			}
		}
	}
	return errors.Join(errs...)
}
