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
	// time: day, hour, minute (lesson 14)
	MsgTime = "Det är dag %d, klockan %02d:%02d."
	// to everyone when the server stops
	MsgShutdown = "Servern stängs av. Vi ses snart!"
	// look: items in the room (lesson 15)
	MsgItemsHere = "Du ser: %s."
	// get without an argument
	MsgGetWhat = "Ta vad?"
	// drop without an argument
	MsgDropWhat = "Släpp vad?"
	// examine without an argument
	MsgExamineWhat = "Undersöka vad?"
	// nothing matches the word here
	MsgNoSuchItem = "Du ser ingen %q här."
	// drop: not in the inventory
	MsgNotCarrying = "Du bär inte på någon %q."
	// get: item name
	MsgYouGet = "Du tar %s."
	// get, to the others: player, item
	MsgGets = "%s tar %s."
	// drop: item name
	MsgYouDrop = "Du släpper %s."
	// drop, to the others: player, item
	MsgDrops = "%s släpper %s."
	// inventory: list of items
	MsgInventory = "Du bär på: %s."
	// inventory: nothing
	MsgInventoryEmpty = "Du bär inte på något."
	// world file: items (lesson 15)
	MsgErrItemFields    = "föremål %q: id och name måste finnas"
	MsgErrDuplicateItem = "föremål %q finns två gånger"
	MsgErrItemRoom      = "föremål %q: rummet %q finns inte"
	// world file: creatures (lesson 16)
	MsgErrDuplicateNPC = "varelse %q finns två gånger"
	MsgErrNPCRoom      = "varelse %q: rummet %q finns inte"
	MsgErrNPCFields    = "varelse %q: id och name måste finnas"
	MsgErrBehavior     = "varelse %q: okänt beteende %q"
	// kill without an argument (lesson 18)
	MsgAttackWhat = "Anfalla vad?"
	// no creature matches
	MsgNoSuchTarget = "Här finns ingen %q att slåss mot."
	// the creature is peaceful
	MsgCantAttack = "Du kan inte anfalla %s."
	// you start a fight: creature
	MsgYouAttack = "Du anfaller %s!"
	// to the others: player, creature
	MsgAttacks = "%s anfaller %s!"
	// your hit: creature, damage
	MsgYouHit = "Du träffar %s (%d i skada)."
	// the creature hits you: creature, damage
	MsgHitsYou = "%s träffar dig (%d i skada)."
	// you won: creature
	MsgYouKill = "Du har besegrat %s!"
	// to the others: player, creature
	MsgKills = "%s har besegrat %s!"
	// you died: the room you wake up in
	MsgYouDie = "Du faller ihop ... och vaknar igen i %s."
	// to the others: player
	MsgDies = "%s faller ihop och försvinner."
	// a creature comes back: creature
	MsgAppears = "%s dyker upp."
	// flee when not in a fight
	MsgNotFighting = "Du slåss inte mot någon."
	// flee with no exits
	MsgFleeNowhere = "Det finns ingenstans att fly!"
	// you fled: direction
	MsgYouFlee = "Du flyr %s!"
	// login with saving (lesson 20)
	MsgAskPassword = "Lösenord:"
	// new player: name
	MsgChoosePassword   = "Välj ett lösenord för %s:"
	MsgPasswordTooShort = "Lösenordet måste vara minst 4 tecken."
	MsgWrongPassword    = "Fel lösenord."
	MsgStoreError       = "Något gick fel med sparningen – försök igen senare."
	// zones (lesson 23)
	MsgBlocked     = "Det går inte att gå dit."
	MsgNoStairs    = "Det finns ingen trappa här."
	MsgGoWhere     = "Gå vart? (t.ex. go north)"
	MsgErrZoneRows = "rum %q: zonen är felaktig (%v) – minst 3×3, lika långa rader och kända tecken"
	MsgErrZoneExit = "rum %q: zonen saknar en öppning för utgången %s"
	// creatures that hunt (lesson 27)
	MsgAttacksYou = "%s anfaller dig!"
	// reloading the world (lesson 31)
	MsgNotAdmin     = "Det får bara administratörer göra."
	MsgNoWorldFile  = "Servern startades utan världsfil (-world) – det finns inget att ladda om."
	MsgReloadFailed = "Världsfilen har fel – ingenting ändrades:\n%v"
	MsgReloaded     = "Världen skimrar till – den har byggts om."
	MsgRoomGone     = "Rummet du var i finns inte längre – du befinner dig någon annanstans."
	// scripts in Lua (lesson 32)
	MsgScriptError      = "Något gick fel i ett skript."
	MsgScriptLoadFailed = "Ett skript har fel:\n%v"
	// plugins in WebAssembly (lesson 33)
	MsgPluginError      = "Något gick fel i ett plugin."
	MsgPluginLoadFailed = "Ett plugin har fel:\n%v"
	// events and hooks (lesson 34)
	MsgCancelled = "Något hindrar dig."
	// copyover (lesson 35)
	MsgCopyoverDone   = "Världen skälver ett ögonblick ... allt är som förut, och ändå nytt."
	MsgCopyoverFailed = "Copyover misslyckades: %v"
	// building in the game (lesson 36)
	MsgDigUsage     = "Så här: @dig <riktning> <namn>, t.ex. @dig north Mörk grotta"
	MsgDescUsage    = "Så här: @desc <beskrivning>"
	MsgNameUsage    = "Så här: @name <namn>"
	MsgExitUsage    = "Så här: @exit <riktning> <rums-id> (eller none för att ta bort)"
	MsgExitExists   = "Det finns redan en utgång åt %s."
	MsgNewRoomDesc  = "Ett tomt rum. Beskriv det med @desc."
	MsgDug          = "Du gräver åt %s och skapar %q (%s)."
	MsgBuilt        = "Klart."
	MsgBuildInvalid = "Det skulle göra världen ogiltig – ingenting ändrades:\n%v"
	MsgNotSaved     = "(Servern har ingen världsfil – ändringen sparas inte.)"
	MsgSaveFailed   = "Världen kunde inte sparas: %v"
	// the tutorial (lesson 37)
	MsgTutorialStep     = "[Introduktion %d/%d] %s"
	MsgTutorialSkipped  = "Du lämnar introduktionen. Skriv help för att se alla kommandon."
	MsgTutorialNotIn    = "Du är inte i introduktionen."
	MsgDisappears       = "%s försvinner."
	MsgErrTutorialStart = "introduktionen: startrummet %q finns inte"
	MsgErrTutorialOpen  = "introduktionen: världens startrum %q går att nå från introduktionen – introduktionens område måste vara avskilt"
	MsgErrTutorialEmpty = "introduktionen: det finns inga steg"
	MsgErrTutorialStep  = "introduktionens steg %d: det behöver en text och precis ett av command, room och item"
	MsgErrTutorialRoom  = "introduktionens steg %d: rummet %q ligger inte i introduktionens område"
	MsgErrTutorialItem  = "introduktionens steg %d: saken %q ligger inte i introduktionens område"
	// languages per player (lesson 38)
	MsgLanguage        = "Ditt språk: %s. Det finns: %s. Välj med t.ex. language sv"
	MsgLanguageSet     = "Språk: %s."
	MsgLanguageUnknown = "Det finns inga texter på %q. Det finns: %s."
	MsgErrText         = "en text måste vara en sträng eller ett objekt med en sträng per språk"
	MsgErrTemplate     = "texten på %q är ingen giltig mall: %v"
	MsgErrTutorialText = "introduktionens steg %d: %v"
	// protection (lesson 43)
	MsgSlowDown           = "Ta det lugnt – du skickar kommandon för fort."
	MsgFlooding           = "För många kommandon – hej då."
	MsgTooManyConnections = "För många anslutningar från din adress – försök igen senare."
	MsgLineTooLong        = "Den raden är för lång – hej då."
	MsgLoginTimeout       = "Det tog för lång tid att logga in – hej då."
)

// CourseLanguage is the language this file – and world.json – is written in (lesson 38).
const CourseLanguage = "sv"

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
