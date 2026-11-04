package mud

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
	"n": "north",
	// TODO(lesson 4): add s, e, w, u, d, l (look), ? (help) and q (quit)
}

// ParseCommand parses a line from the player.
func ParseCommand(line string) Command {
	// TODO(lesson 4)
	return Command{}
}
