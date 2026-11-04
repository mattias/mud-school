package mud

import (
	"fmt"
	"sync"
	"testing"
)

func TestSecondClientIsNotBlockedByFirst(t *testing.T) {
	addr := startServer(t, NewServer())
	idle := connect(t, addr)
	idle.name = "första klienten"
	idle.expect(Welcome)
	// Den första klienten sitter tyst. Den andra ska ändå få svar direkt.

	busy := connect(t, addr)
	busy.name = "andra klienten"
	busy.expect(Welcome)
	busy.send("hallå?")
	busy.expect("hallå?")

	// ... och den första klienten fungerar fortfarande.
	idle.send("jag är kvar")
	idle.expect("jag är kvar")
}

func TestManyClientsAtOnce(t *testing.T) {
	addr := startServer(t, NewServer())
	const n = 50
	clients := make([]*testClient, n)
	for i := range clients {
		clients[i] = connect(t, addr)
		clients[i].name = fmt.Sprintf("klient %d", i)
	}
	for _, c := range clients {
		c.expect(Welcome)
	}

	// Alla skickar samtidigt.
	var wg sync.WaitGroup
	for i, c := range clients {
		wg.Add(1)
		go func(i int, c *testClient) {
			defer wg.Done()
			msg := fmt.Sprintf("meddelande från %d", i)
			fmt.Fprint(c.conn, msg+"\r\n")
		}(i, c)
	}
	wg.Wait()
	for i, c := range clients {
		c.expect(fmt.Sprintf("meddelande från %d", i))
	}
}
