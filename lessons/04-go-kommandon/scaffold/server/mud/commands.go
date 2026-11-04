package mud

// CommandFunc performs a command for a player.
// Functions are ordinary values in Go – they can be stored in maps and passed around.
type CommandFunc func(s *Server, p *Player, cmd Command)

// Register adds (or replaces) a command.
func (s *Server) Register(verb string, fn CommandFunc) {
	s.commands[verb] = fn
}

func (s *Server) registerBuiltins() {
	s.Register("echo", cmdEcho)
	// TODO(lesson 4): register help and quit
}

// dispatch parses a line and runs the right command.
func (s *Server) dispatch(p *Player, line string) {
	// TODO(lesson 4):
	//  - parse the line with ParseCommand
	//  - empty verb: do nothing
	//  - look the verb up in s.commands and call the function
	//  - unknown verb: send MsgUnknownCommand (see messages.go)
}

func cmdEcho(s *Server, p *Player, cmd Command) {
	p.Send(cmd.Rest)
}

// TODO(lesson 4): cmdHelp and cmdQuit
