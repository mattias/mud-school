package mud

// Room is a place in the world.
type Room struct {
	ID          string
	Name        string
	Description string
	Exits       map[string]string // direction ("north") -> room ID
}

// World is all the rooms in the game.
type World struct {
	Start string           // ID of the room where new players appear
	Rooms map[string]*Room // ID -> room
}

// Directions are all the directions in the order they are shown.
// (Their names as shown to players are in messages.go.)
var Directions = []string{"north", "east", "south", "west", "up", "down"}

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
