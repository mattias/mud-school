package mud

import (
	"bytes"
	_ "embed" // behövs för //go:embed
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
)

// Room är en plats i världen. Taggarna `json:"..."` styr hur fälten heter i
// JSON-filen.
type Room struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Terrain     string            `json:"terrain"`         // en av Terrains
	X           int               `json:"x"`               // öster är +1
	Y           int               `json:"y"`               // söder är +1
	Z           int               `json:"z"`               // upp är +1
	Exits       map[string]string `json:"exits,omitempty"` // riktning -> rummets ID
}

// World är alla rum i spelet.
type World struct {
	Start string           // ID för rummet där nya spelare hamnar
	Rooms map[string]*Room // ID -> rum
}

// worldFile är formatet på en världsfil.
type worldFile struct {
	Start string  `json:"start"`
	Rooms []*Room `json:"rooms"`
}

// Directions är alla riktningar i den ordning de ska visas.
var Directions = []string{"north", "east", "south", "west", "up", "down"}

// Terrains är de terrängtyper ett rum kan ha.
var Terrains = []string{"town", "inside", "road", "field", "forest", "water", "mountain", "cave"}

// directionNames är riktningarnas svenska namn, som de visas för spelaren.
var directionNames = map[string]string{
	"north": "norr", "east": "öster", "south": "söder", "west": "väster", "up": "upp", "down": "ner",
}

// directionTravel används när någon går: "Anna går norrut."
var directionTravel = map[string]string{
	"north": "norrut", "east": "österut", "south": "söderut", "west": "västerut", "up": "uppåt", "down": "nedåt",
}

// directionDelta är hur koordinaterna ändras när man går åt ett håll.
var directionDelta = map[string][3]int{
	"north": {0, -1, 0}, "south": {0, 1, 0}, "east": {1, 0, 0}, "west": {-1, 0, 0}, "up": {0, 0, 1}, "down": {0, 0, -1},
}

// NewWorld skapar en värld av rummen. "..." gör funktionen variadisk:
// NewWorld("torg", rum1, rum2, rum3) – rooms blir en []*Room.
func NewWorld(start string, rooms ...*Room) *World {
	w := &World{Start: start, Rooms: make(map[string]*Room, len(rooms))}
	for _, r := range rooms {
		w.Rooms[r.ID] = r
	}
	return w
}

// Room returnerar rummet med givet ID, eller nil om det inte finns.
func (w *World) Room(id string) *Room {
	return w.Rooms[id]
}

// LoadWorld läser och validerar en värld i JSON-format.
func LoadWorld(r io.Reader) (*World, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields() // stavfel som "nmae" ska ge fel, inte ignoreras
	var f worldFile
	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf("läser världsfil: %w", err)
	}
	seen := make(map[string]bool)
	for _, room := range f.Rooms {
		if seen[room.ID] {
			return nil, fmt.Errorf("rum %q finns två gånger", room.ID)
		}
		seen[room.ID] = true
	}
	w := NewWorld(f.Start, f.Rooms...)
	if err := w.Validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// LoadWorldFile läser en värld från en fil.
func LoadWorldFile(path string) (*World, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadWorld(f)
}

// Validate letar efter fel i världen och returnerar ALLA den hittar,
// sammanslagna med errors.Join (eller nil om allt är rätt).
func (w *World) Validate() error {
	var errs []error
	if w.Room(w.Start) == nil {
		errs = append(errs, fmt.Errorf("startrummet %q finns inte", w.Start))
	}

	ids := make([]string, 0, len(w.Rooms))
	for id := range w.Rooms {
		ids = append(ids, id)
	}
	sort.Strings(ids) // stabil ordning på felen

	positions := make(map[[3]int]string)
	for _, id := range ids {
		r := w.Rooms[id]
		if r.ID == "" || r.Name == "" {
			errs = append(errs, fmt.Errorf("rum %q: id och name måste finnas", id))
		}
		if !slices.Contains(Terrains, r.Terrain) {
			errs = append(errs, fmt.Errorf("rum %q: okänd terräng %q", id, r.Terrain))
		}
		pos := [3]int{r.X, r.Y, r.Z}
		if other, taken := positions[pos]; taken {
			errs = append(errs, fmt.Errorf("rum %q: har samma koordinater %v som %q", id, pos, other))
		}
		positions[pos] = id

		for _, dir := range Directions {
			to, ok := r.Exits[dir]
			if !ok {
				continue
			}
			target := w.Room(to)
			if target == nil {
				errs = append(errs, fmt.Errorf("rum %q: utgången %s leder till %q som inte finns", id, dir, to))
				continue
			}
			d := directionDelta[dir]
			if target.X != r.X+d[0] || target.Y != r.Y+d[1] || target.Z != r.Z+d[2] {
				errs = append(errs, fmt.Errorf("rum %q: utgången %s leder till %q, men koordinaterna passar inte ihop", id, dir, to))
			}
		}
		for dir := range r.Exits {
			if _, ok := directionDelta[dir]; !ok {
				errs = append(errs, fmt.Errorf("rum %q: okänd riktning %q", id, dir))
			}
		}
	}
	return errors.Join(errs...)
}

//go:embed world.json
var defaultWorldJSON []byte

// DefaultWorld är världen i world.json, som är inbakad i programfilen.
func DefaultWorld() *World {
	w, err := LoadWorld(bytes.NewReader(defaultWorldJSON))
	if err != nil {
		panic("world.json är trasig: " + err.Error())
	}
	return w
}
