package mud

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

// outboxSize är hur många meddelanden som får vänta på att skickas till en
// spelare innan vi anser att klienten är för långsam.
const outboxSize = 256

// Player är en ansluten spelare.
type Player struct {
	Name string

	conn    net.Conn
	scanner *bufio.Scanner
	quit    bool // sätts till true när spelaren vill sluta

	outbox chan string   // allt som ska skickas till spelaren
	done   chan struct{} // stängs när skrivargoroutinen är klar
}

func newPlayer(conn net.Conn) *Player {
	p := &Player{
		conn:    conn,
		scanner: bufio.NewScanner(conn),
		outbox:  make(chan string, outboxSize),
		done:    make(chan struct{}),
	}
	go p.writeLoop()
	return p
}

// writeLoop är den ENDA goroutine som skriver till p.conn.
func (p *Player) writeLoop() {
	defer close(p.done)
	for msg := range p.outbox {
		if _, err := io.WriteString(p.conn, msg); err != nil {
			// Anslutningen är död. Fortsätt tömma kanalen så att ingen fastnar.
			continue
		}
	}
}

// Send köar text till spelaren. Send blockerar aldrig: är utkorgen full
// kopplas spelaren ner istället.
func (p *Player) Send(text string) {
	msg := strings.ReplaceAll(text, "\n", "\r\n") + "\r\n"
	select {
	case p.outbox <- msg:
	default:
		log.Printf("%s är för långsam – kopplar ner", p.conn.RemoteAddr())
		p.conn.Close()
	}
}

// Sendf är som Send men med fmt.Sprintf-formatering.
func (p *Player) Sendf(format string, args ...any) {
	p.Send(fmt.Sprintf(format, args...))
}

// readLine läser nästa rad från spelaren. ok är false om spelaren kopplat ner.
func (p *Player) readLine() (line string, ok bool) {
	if !p.scanner.Scan() {
		return "", false
	}
	return p.scanner.Text(), true
}

// close stänger utkorgen och väntar tills allt i den har skickats.
// Efter close får ingen anropa Send.
func (p *Player) close() {
	close(p.outbox)
	<-p.done
}
