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
// game loop itself – a server whose loop is stuck is not healthy either. The
// uptime counts from when the handler was made, which is when main starts.
func (s *Server) HealthHandler() http.Handler {
	started := time.Now()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := Health{Status: "stopping"}
		code := http.StatusServiceUnavailable
		if !s.closing.Load() {
			// Not s.do: it waits for as long as the loop takes to take the work.
			// Every wait here has a time limit, so a stuck loop gives a 503.
			limit := time.After(2 * time.Second)
			done := make(chan int, 1)
			s.start()
			select {
			case s.events <- func() { done <- len(s.players) }:
				select {
				case n := <-done:
					h = Health{Status: "ok", Players: n, Uptime: int64(time.Since(started).Seconds())}
					code = http.StatusOK
				case <-limit:
				case <-r.Context().Done():
				}
			case <-limit: // the loop doesn't even take the question
			case <-s.loopDone:
			case <-r.Context().Done():
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(h)
	})
}
