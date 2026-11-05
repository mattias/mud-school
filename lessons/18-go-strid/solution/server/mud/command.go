package mud

import "strings"

// Command is a parsed line from the player.
//
//	"Get Sword from chest" -> Command{Verb: "get", Args: []string{"Sword", "from", "chest"}, Rest: "Sword from chest"}
type Command struct {
	Verb string   // the first word, lower case, with abbreviations expanded
	Args []string // the remaining words
	Rest string   // the rest of the line as typed (trimmed)
}

// aliases maps abbreviations to full verbs.
var aliases = map[string]string{
	"n":      "north",
	"s":      "south",
	"e":      "east",
	"w":      "west",
	"u":      "up",
	"d":      "down",
	"l":      "look",
	"?":      "help",
	"q":      "quit",
	"i":      "inventory",
	"x":      "examine",
	"take":   "get",
	"k":      "kill",
	"attack": "kill",
}

// ParseCommand parses a line from the player.
func ParseCommand(line string) Command {
	line = strings.TrimSpace(line)
	if line == "" {
		return Command{}
	}
	// 'hello  ==  say hello  (a classic MUD shortcut)
	if rest, ok := strings.CutPrefix(line, "'"); ok {
		rest = strings.TrimSpace(rest)
		return Command{Verb: "say", Args: strings.Fields(rest), Rest: rest}
	}
	verb, rest, _ := strings.Cut(line, " ")
	verb = strings.ToLower(verb)
	if full, ok := aliases[verb]; ok {
		verb = full
	}
	rest = strings.TrimSpace(rest)
	return Command{Verb: verb, Args: strings.Fields(rest), Rest: rest}
}
