// Code generated for the course language: English.
//
// All text the game shows to players lives in this file, so that the course
// can be taken in Swedish, English or Japanese. The tests use these constants
// too, so keep the names – but feel free to change the texts.
// `./mudschool lang <en|sv|ja>` replaces this file.
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
	// time: day, hour, minute (lesson 14)
	MsgTime = "It is day %d, %02d:%02d."
	// to everyone when the server stops
	MsgShutdown = "The server is shutting down. See you soon!"
	// look: items in the room (lesson 15)
	MsgItemsHere = "You see: %s."
	// get without an argument
	MsgGetWhat = "Get what?"
	// drop without an argument
	MsgDropWhat = "Drop what?"
	// examine without an argument
	MsgExamineWhat = "Examine what?"
	// nothing matches the word here
	MsgNoSuchItem = "You don't see %q here."
	// drop: not in the inventory
	MsgNotCarrying = "You aren't carrying %q."
	// get: item name
	MsgYouGet = "You pick up %s."
	// get, to the others: player, item
	MsgGets = "%s picks up %s."
	// drop: item name
	MsgYouDrop = "You drop %s."
	// drop, to the others: player, item
	MsgDrops = "%s drops %s."
	// inventory: list of items
	MsgInventory = "You are carrying: %s."
	// inventory: nothing
	MsgInventoryEmpty = "You aren't carrying anything."
	// world file: items (lesson 15)
	MsgErrItemFields    = "item %q: id and name are required"
	MsgErrDuplicateItem = "item %q appears twice"
	MsgErrItemRoom      = "item %q: room %q does not exist"
	// world file: creatures (lesson 16)
	MsgErrDuplicateNPC = "creature %q appears twice"
	MsgErrNPCRoom      = "creature %q: room %q does not exist"
	MsgErrNPCFields    = "creature %q: id and name are required"
	MsgErrBehavior     = "creature %q: unknown behaviour %q"
	// kill without an argument (lesson 18)
	MsgAttackWhat = "Attack what?"
	// no creature matches
	MsgNoSuchTarget = "There is no %q here to fight."
	// the creature is peaceful
	MsgCantAttack = "You can't attack %s."
	// you start a fight: creature
	MsgYouAttack = "You attack %s!"
	// to the others: player, creature
	MsgAttacks = "%s attacks %s!"
	// your hit: creature, damage
	MsgYouHit = "You hit %s (%d damage)."
	// the creature hits you: creature, damage
	MsgHitsYou = "%s hits you (%d damage)."
	// you won: creature
	MsgYouKill = "You have defeated %s!"
	// to the others: player, creature
	MsgKills = "%s has defeated %s!"
	// you died: the room you wake up in
	MsgYouDie = "You collapse ... and wake up again in %s."
	// to the others: player
	MsgDies = "%s collapses and vanishes."
	// a creature comes back: creature
	MsgAppears = "%s appears."
	// flee when not in a fight
	MsgNotFighting = "You aren't fighting anyone."
	// flee with no exits
	MsgFleeNowhere = "There is nowhere to flee!"
	// you fled: direction
	MsgYouFlee = "You flee %s!"
	// login with saving (lesson 20)
	MsgAskPassword = "Password:"
	// new player: name
	MsgChoosePassword   = "Choose a password for %s:"
	MsgPasswordTooShort = "The password must be at least 4 characters."
	MsgWrongPassword    = "Wrong password."
	MsgStoreError       = "Something went wrong with saving – please try again later."
	// zones (lesson 23)
	MsgBlocked     = "You can't walk there."
	MsgNoStairs    = "There are no stairs here."
	MsgGoWhere     = "Go where? (e.g. go north)"
	MsgErrZoneRows = "room %q: the zone is invalid (%v) – at least 3×3, rows of equal length and known tiles"
	MsgErrZoneExit = "room %q: the zone has no opening for the exit %s"
	// creatures that hunt (lesson 27)
	MsgAttacksYou = "%s attacks you!"
	// reloading the world (lesson 31)
	MsgNotAdmin     = "Only administrators can do that."
	MsgNoWorldFile  = "The server was started without a world file (-world) – there is nothing to reload."
	MsgReloadFailed = "The world file has errors – nothing was changed:\n%v"
	MsgReloaded     = "The world shimmers – it has been rebuilt."
	MsgRoomGone     = "The room you were in is gone – you find yourself somewhere else."
	// scripts in Lua (lesson 32)
	MsgScriptError      = "Something went wrong in a script."
	MsgScriptLoadFailed = "A script has errors:\n%v"
	// plugins in WebAssembly (lesson 33)
	MsgPluginError      = "Something went wrong in a plugin."
	MsgPluginLoadFailed = "A plugin has errors:\n%v"
	// events and hooks (lesson 34)
	MsgCancelled = "Something stops you."
	// copyover (lesson 35)
	MsgCopyoverDone   = "The world shudders for a moment ... everything is as it was, and yet new."
	MsgCopyoverFailed = "Copyover failed: %v"
	// building in the game (lesson 36)
	MsgDigUsage     = "Like this: @dig <direction> <name>, e.g. @dig north Dark cave"
	MsgDescUsage    = "Like this: @desc <description>"
	MsgNameUsage    = "Like this: @name <name>"
	MsgExitUsage    = "Like this: @exit <direction> <room id> (or none to remove it)"
	MsgExitExists   = "There is already an exit %s."
	MsgNewRoomDesc  = "An empty room. Describe it with @desc."
	MsgDug          = "You dig %s and create %q (%s)."
	MsgBuilt        = "Done."
	MsgBuildInvalid = "That would make the world invalid – nothing was changed:\n%v"
	MsgNotSaved     = "(The server has no world file – the change is not saved.)"
	MsgSaveFailed   = "The world could not be saved: %v"
	// the tutorial (lesson 37)
	MsgTutorialStep     = "[Tutorial %d/%d] %s"
	MsgTutorialSkipped  = "You leave the tutorial. Type help to see every command."
	MsgTutorialNotIn    = "You are not in the tutorial."
	MsgDisappears       = "%s disappears."
	MsgErrTutorialStart = "tutorial: the start room %q doesn't exist"
	MsgErrTutorialOpen  = "tutorial: the world's start room %q can be reached from the tutorial – the tutorial area must be closed off"
	MsgErrTutorialEmpty = "tutorial: there are no steps"
	MsgErrTutorialStep  = "tutorial step %d: it needs a text and exactly one of command, room and item"
	MsgErrTutorialRoom  = "tutorial step %d: the room %q is not in the tutorial area"
	MsgErrTutorialItem  = "tutorial step %d: the item %q doesn't lie in the tutorial area"
	// languages per player (lesson 38)
	MsgLanguage        = "Your language: %s. There are: %s. Choose with e.g. language en"
	MsgLanguageSet     = "Language: %s."
	MsgLanguageUnknown = "There are no texts in %q. There are: %s."
	MsgErrText         = "a text must be a string or an object with one string per language"
	MsgErrTemplate     = "the text in %q is not a valid template: %v"
	MsgErrTutorialText = "tutorial step %d: %v"
	// protection (lesson 43)
	MsgSlowDown           = "Slow down – you are sending commands too fast."
	MsgFlooding           = "Too many commands – goodbye."
	MsgTooManyConnections = "Too many connections from your address – try again later."
	MsgLineTooLong        = "That line is too long – goodbye."
	MsgLoginTimeout       = "You took too long to log in – goodbye."
)

// CourseLanguage is the language this file – and world.json – is written in (lesson 38).
const CourseLanguage = "en"

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
