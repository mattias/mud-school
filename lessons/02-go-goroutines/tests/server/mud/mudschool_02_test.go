package mud

import (
	"fmt"
	"sync"
	"testing"
)

func TestSecondClientIsNotBlockedByFirst(t *testing.T) {
	addr := startServer(t, NewServer())
	idle := connect(t, addr)
	idle.name = L("första klienten", "first client", "1 番目のクライアント")
	idle.expect(Welcome)
	// The first client stays silent. The second one must still get answers right away.

	busy := connect(t, addr)
	busy.name = L("andra klienten", "second client", "2 番目のクライアント")
	busy.expect(Welcome)
	busy.send("hello?")
	busy.expect("hello?")

	// ... and the first client still works.
	idle.send("still here")
	idle.expect("still here")
}

func TestManyClientsAtOnce(t *testing.T) {
	addr := startServer(t, NewServer())
	const n = 50
	clients := make([]*testClient, n)
	for i := range clients {
		clients[i] = connect(t, addr)
		clients[i].name = fmt.Sprintf("#%d", i)
	}
	for _, c := range clients {
		c.expect(Welcome)
	}

	// Everybody sends at the same time.
	var wg sync.WaitGroup
	for i, c := range clients {
		wg.Add(1)
		go func(i int, c *testClient) {
			defer wg.Done()
			fmt.Fprintf(c.conn, "message from %d\r\n", i)
		}(i, c)
	}
	wg.Wait()
	for i, c := range clients {
		c.expect(fmt.Sprintf("message from %d", i))
	}
}
