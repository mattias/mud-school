package mud

// One tick is one minute of game time. The game starts on day 1 at 08:00.
const startMinute = 8 * 60

// GameTime converts a number of ticks to the game's day, hour and minute.
func GameTime(ticks int64) (day, hour, minute int) {
	// TODO(lesson 14)
	return 1, 8, 0
}

// cmdTime shows the game time (register it as "time"). Runs in the game loop.
func cmdTime(s *Server, p *Player, cmd Command) {
	day, hour, minute := GameTime(s.ticks)
	p.Sendf(MsgTime, day, hour, minute)
}
