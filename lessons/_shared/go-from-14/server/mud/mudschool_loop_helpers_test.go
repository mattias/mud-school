// This file belongs to the course tool and is overwritten by `mudschool check`.
// Helpers for lesson 14 and later (they need Server.Shutdown).
package mud

import (
	"context"
	"testing"
	"time"
)

// shutdown stops a server and fails the test if it doesn't stop cleanly.
func shutdown(t *testing.T, srv *Server) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}
