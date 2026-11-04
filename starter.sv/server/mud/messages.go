// Code generated for the course language: svenska.
//
// All text the game shows to players lives in this file, so that the course
// can be taken in Swedish, English or Japanese. The tests use these constants
// too, so keep the names – but feel free to change the texts.
// `./mudschool lang <en|sv|ja>` replaces this file.
package mud

const (
	// the first thing a new player sees (lesson 1)
	Welcome = "Välkommen till MUD School!"
	// login prompt (lesson 5)
	MsgAskName = "Vad heter du?"
	// after login: name
	MsgWelcomeName = "Välkommen, %s!"
	// bad name: error
	MsgInvalidName     = "Ogiltigt namn: %v."
	MsgErrNameTooShort = "namnet måste vara minst 2 tecken"
	MsgErrNameTooLong  = "namnet får vara högst 16 tecken"
	MsgErrNameInvalid  = "namnet får bara innehålla bokstäver"
	MsgErrNameTaken    = "det namnet är redan upptaget"
	// unknown verb, formatted with %q (lesson 4)
	MsgUnknownCommand = "Okänt kommando %q. Skriv help för hjälp."
	// help: comma separated verbs
	MsgCommands = "Kommandon: %s"
	// quit
	MsgGoodbye = "Hej då!"
	// who: count, names (lesson 5)
	MsgWho = "Spelare online (%d): %s"
	// to everyone else when a player logs in (lesson 7)
	MsgJoined = "%s har anslutit."
	// to everyone else when a player leaves
	MsgLeft = "%s har lämnat."
	// say, to the speaker
	MsgYouSay = "Du säger: %s"
	// say, to the others: name, text
	MsgSays = "%s säger: %s"
	// say without text
	MsgSayWhat = "Säg vad?"
	// shout, to the shouter (lesson 8)
	MsgYouShout = "Du ropar: %s"
	// shout, to everyone else: name, text
	MsgShouts = "%s ropar: %s"
	// shout without text
	MsgShoutWhat = "Ropa vad?"
	// look: list of exits
	MsgExits = "Utgångar: %s"
	// look: no exits
	MsgNoExits = "Utgångar: inga"
	// look: other players in the room
	MsgHere = "Här finns: %s"
	// no exit in that direction
	MsgCantGo = "Du kan inte gå åt det hållet."
	// to the old room: name, directionTravel[dir]
	MsgLeaves = "%s går %s."
	// to the new room: name
	MsgArrives = "%s kommer in."
	// separator in lists of names and exits
	MsgListSep = ", "
	// world file errors (lesson 9) – must keep %w
	MsgErrReadWorld     = "läser världsfil: %w"
	MsgErrDuplicateRoom = "rum %q finns två gånger"
	MsgErrStartMissing  = "startrummet %q finns inte"
	MsgErrRoomFields    = "rum %q: id och name måste finnas"
	MsgErrTerrain       = "rum %q: okänd terräng %q"
	MsgErrSamePos       = "rum %q: har samma koordinater %v som %q"
	MsgErrExitMissing   = "rum %q: utgången %s leder till %q som inte finns"
	MsgErrExitCoords    = "rum %q: utgången %s leder till %q, men koordinaterna passar inte ihop"
	MsgErrDirection     = "rum %q: okänd riktning %q"
)

// directionNames are the directions as shown to the player (lesson 8).
var directionNames = map[string]string{
	"north": "norr",
	"east":  "öster",
	"south": "söder",
	"west":  "väster",
	"up":    "upp",
	"down":  "ner",
}

// directionTravel is used in MsgLeaves (lesson 8).
var directionTravel = map[string]string{
	"north": "norrut",
	"east":  "österut",
	"south": "söderut",
	"west":  "västerut",
	"up":    "uppåt",
	"down":  "nedåt",
}
