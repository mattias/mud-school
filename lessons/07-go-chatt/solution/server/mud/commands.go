package mud

import (
	"fmt"
	"sort"
	"strings"
)

// CommandFunc performs a command for a player.
// Functions are ordinary values in Go – they can be stored in maps and passed around.
type CommandFunc func(s *Server, p *Player, cmd Command)

// Register adds (or replaces) a command.
func (s *Server) Register(verb string, fn CommandFunc) {
	s.commands[verb] = fn
}

func (s *Server) registerBuiltins() {
	s.Register("echo", cmdEcho)
	s.Register("help", cmdHelp)
	s.Register("quit", cmdQuit)
	s.Register("who", cmdWho)
	s.Register("say", cmdSay)
}

// dispatch parses a line and runs the right command.
func (s *Server) dispatch(p *Player, line string) {
	cmd := ParseCommand(line)
	if cmd.Verb == "" {
		return
	}
	fn, ok := s.commands[cmd.Verb]
	if !ok {
		p.Sendf(MsgUnknownCommand, cmd.Verb)
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
	p.Sendf(MsgCommands, strings.Join(verbs, MsgListSep))
}

func cmdQuit(s *Server, p *Player, cmd Command) {
	p.Send(MsgGoodbye)
	p.quit = true
}

func cmdWho(s *Server, p *Player, cmd Command) {
	names := s.PlayerNames()
	p.Sendf(MsgWho, len(names), strings.Join(names, MsgListSep))
}

func cmdSay(s *Server, p *Player, cmd Command) {
	if cmd.Rest == "" {
		p.Send(MsgSayWhat)
		return
	}
	p.Sendf(MsgYouSay, cmd.Rest)
	s.broadcast(p, fmt.Sprintf(MsgSays, p.Name, cmd.Rest))
}
