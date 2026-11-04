package mud

// CommandFunc är en funktion som utför ett kommando åt en spelare.
// Funktioner är vanliga värden i Go – de kan läggas i maps och skickas runt.
type CommandFunc func(s *Server, p *Player, cmd Command)

// Register lägger till (eller ersätter) ett kommando.
func (s *Server) Register(verb string, fn CommandFunc) {
	s.commands[verb] = fn
}

func (s *Server) registerBuiltins() {
	s.Register("echo", cmdEcho)
	// TODO(lektion 4): registrera help och quit
}

// dispatch tolkar en rad och kör rätt kommando.
func (s *Server) dispatch(p *Player, line string) {
	// TODO(lektion 4):
	//  - tolka raden med ParseCommand
	//  - tomt verb: gör ingenting
	//  - leta upp verbet i s.commands och kör funktionen
	//  - okänt verb: skicka `Okänt kommando "<verb>". Skriv help för hjälp.`
}

func cmdEcho(s *Server, p *Player, cmd Command) {
	p.Send(cmd.Rest)
}

// TODO(lektion 4): cmdHelp och cmdQuit
