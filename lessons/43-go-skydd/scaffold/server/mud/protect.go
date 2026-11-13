package mud

import (
	"expvar"
	"net"
	"time"
)

// Protection (lesson 43): on the open internet somebody will paste a whole
// book into your MUD, open a thousand connections or just connect and never
// say anything. None of that may slow the game down for everybody else.

// DefaultMaxLine is the longest line a player may send if
// Options.MaxLineLength is 0.
const DefaultMaxLine = 4096

// FloodLimit is how many commands in a row may be dropped by the rate
// limiter before the player is disconnected.
const FloodLimit = 100

// RateLimiter is a token bucket: it holds up to burst tokens and fills up
// with rate tokens per second. Every command costs one token; with none
// left the command is dropped. Short bursts are fine, a steady flood is not.
type RateLimiter struct {
	rate, burst float64
	tokens      float64
	last        time.Time
}

// NewRateLimiter creates a full bucket.
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	return &RateLimiter{rate: rate, burst: float64(burst), tokens: float64(burst)}
}

// Allow takes a token at time now, if there is one. The time is passed in,
// so the tests don't have to wait.
func (r *RateLimiter) Allow(now time.Time) bool {
	// TODO(lesson 43): add (now - r.last) × rate tokens (not on the first
	// call), at most burst; remember now; take one token if there is one.
	return true
}

// The server's numbers, for /debug/vars (expvar). They are counted for the
// whole program, so they are created once, here.
var (
	metricConnections = expvar.NewInt("mud_connections_total")
	metricRefused     = expvar.NewInt("mud_connections_refused_total")
	metricCommands    = expvar.NewInt("mud_commands_total")
	metricDropped     = expvar.NewInt("mud_commands_dropped_total")
	metricPlayers     = expvar.NewInt("mud_players")
)

// hostOf is the IP address of a connection's other end, without the port.
func hostOf(a net.Addr) string {
	host, _, err := net.SplitHostPort(a.String())
	if err != nil {
		return a.String()
	}
	return host
}

// connsFrom counts the connections from ip. Runs in the game loop.
func (s *Server) connsFrom(ip string) int {
	// TODO(lesson 43): count the connections in s.conns with p.ip == ip.
	return 0
}

// allowCommand runs the player's rate limiter. It returns false if the
// command should be dropped, and quit = true if the player floods.
func (s *Server) allowCommand(p *Player) (ok, quit bool) {
	// TODO(lesson 43): allowed: dropped = 0, count the command. Otherwise count
	// the drop, p.dropped++, MsgSlowDown on the first drop only, and
	// MsgFlooding + quit at FloodLimit.
	return true, false
}
