package mud

import (
	"encoding/json"
	"net/http"
	"time"
)

// Health (lesson 42): a page that says whether the server is alive, for
// Docker's HEALTHCHECK, a load balancer or a monitoring service.

// Health is what /healthz answers.
type Health struct {
	Status  string `json:"status"` // "ok" or "stopping"
	Players int    `json:"players"`
	Uptime  int64  `json:"uptime_seconds"`
}

// HealthHandler answers GET /healthz with 200 and a Health while the game
// loop runs, and 503 when the server is stopping or has stopped. It asks the
// game loop itself – a server whose loop is stuck is not healthy either.
func (s *Server) HealthHandler() http.Handler {
	// TODO(lesson 42): a handler that asks the game loop (s.do, with a time
	// limit) for the number of players: 200 {status: ok, players, uptime_seconds};
	// while closing, or if the loop doesn't answer: 503 {status: stopping}.
	_, _ = json.Marshal, time.Now
	return http.NotFoundHandler()
}
