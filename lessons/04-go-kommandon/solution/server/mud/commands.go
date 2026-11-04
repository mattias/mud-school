package mud

import (
	"sort"
	"strings"
)

// CommandFunc är en funktion som utför ett kommando åt en spelare.
// Funktioner är vanliga värden i Go – de kan läggas i maps och skickas runt.
type CommandFunc func(s *Server, p *Player, cmd Command)

// Register lägger till (eller ersätter) ett kommando.
func (s *Server) Register(verb string, fn CommandFunc) {
	s.commands[verb] = fn
}

func (s *Server) registerBuiltins() {
	s.Register("echo", cmdEcho)
	s.Register("help", cmdHelp)
	s.Register("quit", cmdQuit)
}

// dispatch tolkar en rad och kör rätt kommando.
func (s *Server) dispatch(p *Player, line string) {
	cmd := ParseCommand(line)
	if cmd.Verb == "" {
		return
	}
	fn, ok := s.commands[cmd.Verb]
	if !ok {
		p.Sendf("Okänt kommando %q. Skriv help för hjälp.", cmd.Verb)
		return
	}
	fn(s, p, cmd)
}

func cmdEcho(s *Server, p *Player, cmd Command) {
	p.Send(cmd.Rest)
}

func cmdHelp(s *Server, p *Player, cmd Command) {
	verbs := make([]string, 0, len(s.commands))
	for verb := range s.commands {
		verbs = append(verbs, verb)
	}
	sort.Strings(verbs)
	p.Send("Kommandon: " + strings.Join(verbs, ", "))
}

func cmdQuit(s *Server, p *Player, cmd Command) {
	p.Send("Hej då!")
	p.quit = true
}
