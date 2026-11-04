package mud

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
	// TODO(lektion 4): lägg till s, e, w, u, d, l (look), ? (help) och q (quit)
}

// ParseCommand tolkar en rad från spelaren.
func ParseCommand(line string) Command {
	// TODO(lektion 4)
	return Command{}
}
