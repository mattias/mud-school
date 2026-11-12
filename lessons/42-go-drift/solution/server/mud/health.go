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
	started := time.Now()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := Health{Status: "stopping"}
		code := http.StatusServiceUnavailable
		if !s.closing.Load() {
			done := make(chan int, 1)
			if s.do(func() { done <- len(s.players) }) {
				select {
				case n := <-done:
					h = Health{Status: "ok", Players: n, Uptime: int64(time.Since(started).Seconds())}
					code = http.StatusOK
				case <-time.After(2 * time.Second): // the loop doesn't answer
				case <-r.Context().Done():
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(h)
	})
}
