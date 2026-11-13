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
	s.Register("shout", cmdShout)
	s.Register("look", cmdLook)
	s.Register("map", cmdMap)
	s.Register("time", cmdTime)
	s.Register("get", cmdGet)
	s.Register("drop", cmdDrop)
	s.Register("inventory", cmdInventory)
	s.Register("examine", cmdExamine)
	s.Register("kill", cmdKill)
	s.Register("flee", cmdFlee)
	for _, dir := range StepDirections {
		s.Register(dir, stepCommand(dir))
	}
	s.Register("up", stairsCommand("up"))
	s.Register("down", stairsCommand("down"))
	s.Register("go", cmdGo)
	s.Register("zone", cmdZone)
	s.Register("@reload", adminOnly(cmdReload))
	s.Register("@copyover", adminOnly(cmdCopyover))
	s.Register("@dig", adminOnly(cmdDig))
	s.Register("@desc", adminOnly(cmdDesc))
	s.Register("@name", adminOnly(cmdName))
	s.Register("@exit", adminOnly(cmdExit))
	s.Register("tutorial", cmdTutorial) // lesson 37
}

// dispatch parses a line and runs the right command.
func (s *Server) dispatch(p *Player, line string) {
	cmd := ParseCommand(line)
	if cmd.Verb == "" {
		return
	}
	// Built-in commands first – extensions can add commands, not replace them.
	if fn, ok := s.commands[cmd.Verb]; ok {
		fn(s, p, cmd)
	} else if !s.runScriptCommand(p, cmd) && !s.runPluginCommand(p, cmd) { // lessons 32, 33
		p.SendKind(KindError, fmt.Sprintf(MsgUnknownCommand, cmd.Verb))
		return
	}
	// Tell the hooks that a command has run (lesson 37). p may have quit or
	// moved; Room is where they are now.
	s.emit(&Event{Kind: EventCommand, Player: p, Room: p.Room, Text: cmd.Verb})
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
	names := s.playerNames()
	p.Sendf(MsgWho, len(names), strings.Join(names, MsgListSep))
}
