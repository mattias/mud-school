package mud

import "strings"

// Command är en tolkad rad från spelaren.
//
//	"Get Sword from chest" -> Command{Verb: "get", Args: []string{"Sword", "from", "chest"}, Rest: "Sword from chest"}
type Command struct {
	Verb string   // första ordet, med små bokstäver och förkortningar expanderade
	Args []string // resten av orden
	Rest string   // resten av raden som den skrevs (utan mellanslag i början/slutet)
}

// aliases mappar förkortningar till fullständiga verb.
var aliases = map[string]string{
	"n": "north",
	"s": "south",
	"e": "east",
	"w": "west",
	"u": "up",
	"d": "down",
	"l": "look",
	"?": "help",
	"q": "quit",
}

// ParseCommand tolkar en rad från spelaren.
func ParseCommand(line string) Command {
	line = strings.TrimSpace(line)
	if line == "" {
		return Command{}
	}
	// 'hej  ==  say hej  (klassisk MUD-genväg)
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
