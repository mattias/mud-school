// Code generated for the course language: English.
//
// All text the game shows to players lives in this file, so that the course
// can be taken in Swedish, English or Japanese. The tests use these constants
// too, so keep the names – but feel free to change the texts.
// `./mudschool lang <sv|en|ja>` replaces this file.
package mud

const (
	// the first thing a new player sees (lesson 1)
	Welcome = "Welcome to MUD School!"
	// login prompt (lesson 5)
	MsgAskName = "What is your name?"
	// after login: name
	MsgWelcomeName = "Welcome, %s!"
	// bad name: error
	MsgInvalidName     = "Invalid name: %v."
	MsgErrNameTooShort = "the name must be at least 2 characters"
	MsgErrNameTooLong  = "the name may be at most 16 characters"
	MsgErrNameInvalid  = "the name may only contain letters"
	MsgErrNameTaken    = "that name is already taken"
	// unknown verb, formatted with %q (lesson 4)
	MsgUnknownCommand = "Unknown command %q. Type help for help."
	// help: comma separated verbs
	MsgCommands = "Commands: %s"
	// quit
	MsgGoodbye = "Goodbye!"
	// who: count, names (lesson 5)
	MsgWho = "Players online (%d): %s"
	// to everyone else when a player logs in (lesson 7)
	MsgJoined = "%s has connected."
	// to everyone else when a player leaves
	MsgLeft = "%s has left."
	// say, to the speaker
	MsgYouSay = "You say: %s"
	// say, to the others: name, text
	MsgSays = "%s says: %s"
	// say without text
	MsgSayWhat = "Say what?"
	// shout, to the shouter (lesson 8)
	MsgYouShout = "You shout: %s"
	// shout, to everyone else: name, text
	MsgShouts = "%s shouts: %s"
	// shout without text
	MsgShoutWhat = "Shout what?"
	// look: list of exits
	MsgExits = "Exits: %s"
	// look: no exits
	MsgNoExits = "Exits: none"
	// look: other players in the room
	MsgHere = "Also here: %s"
	// no exit in that direction
	MsgCantGo = "You can't go that way."
	// to the old room: name, directionTravel[dir]
	MsgLeaves = "%s leaves %s."
	// to the new room: name
	MsgArrives = "%s arrives."
	// separator in lists of names and exits
	MsgListSep = ", "
	// world file errors (lesson 9) – must keep %w
	MsgErrReadWorld     = "reading world file: %w"
	MsgErrDuplicateRoom = "room %q appears twice"
	MsgErrStartMissing  = "the start room %q does not exist"
	MsgErrRoomFields    = "room %q: id and name are required"
	MsgErrTerrain       = "room %q: unknown terrain %q"
	MsgErrSamePos       = "room %q: has the same coordinates %v as %q"
	MsgErrExitMissing   = "room %q: exit %s leads to %q, which does not exist"
	MsgErrExitCoords    = "room %q: exit %s leads to %q, but the coordinates do not match"
	MsgErrDirection     = "room %q: unknown direction %q"
)

// directionNames are the directions as shown to the player (lesson 8).
var directionNames = map[string]string{
	"north": "north",
	"east":  "east",
	"south": "south",
	"west":  "west",
	"up":    "up",
	"down":  "down",
}

// directionTravel is used in MsgLeaves (lesson 8).
var directionTravel = map[string]string{
	"north": "north",
	"east":  "east",
	"south": "south",
	"west":  "west",
	"up":    "up",
	"down":  "down",
}
