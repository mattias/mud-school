package mud

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync/atomic"
)

// outboxSize är hur många meddelanden som får vänta på att skickas till en
// spelare innan vi anser att klienten är för långsam.
const outboxSize = 256

// Player är en ansluten spelare.
type Player struct {
	Name string
	Room *Room // var spelaren står; skyddas av Server.mu

	conn    net.Conn
	scanner *bufio.Scanner
	quit    bool // sätts till true när spelaren vill sluta

	structured atomic.Bool   // true = klienten vill ha JSON (se HelloPrefix)
	outbox     chan Frame    // allt som ska skickas till spelaren
	done       chan struct{} // stängs när skrivargoroutinen är klar
}

func newPlayer(conn net.Conn) *Player {
	p := &Player{
		conn:    conn,
		scanner: bufio.NewScanner(conn),
		outbox:  make(chan Frame, outboxSize),
		done:    make(chan struct{}),
	}
	go p.writeLoop()
	return p
}

// writeLoop är den ENDA goroutine som skriver till p.conn. Den bestämmer
// också hur varje Frame ska kodas: som text eller som JSON.
func (p *Player) writeLoop() {
	defer close(p.done)
	for f := range p.outbox {
		var wire string
		if p.structured.Load() {
			wire = encodeJSON(f)
		} else {
			wire = encodeText(f)
		}
		if wire == "" {
			continue
		}
		if _, err := io.WriteString(p.conn, wire); err != nil {
			continue // anslutningen är död; töm kanalen så att ingen fastnar
		}
	}
}

// SendFrame köar en Frame till spelaren. Den blockerar aldrig: är utkorgen
// full kopplas spelaren ner istället.
func (p *Player) SendFrame(f Frame) {
	select {
	case p.outbox <- f:
	default:
		log.Printf("%s är för långsam – kopplar ner", p.conn.RemoteAddr())
		p.conn.Close()
	}
}

// Send skickar text till spelaren.
func (p *Player) Send(text string) {
	p.SendFrame(TextFrame(text))
}

// Sendf är som Send men med fmt.Sprintf-formatering.
func (p *Player) Sendf(format string, args ...any) {
	p.Send(fmt.Sprintf(format, args...))
}

// Structured talar om ifall klienten vill ha JSON.
func (p *Player) Structured() bool {
	return p.structured.Load()
}

// readLine läser nästa rad från spelaren. ok är false om spelaren kopplat ner.
// Protokollrader (som "@hello") hanteras här och syns aldrig för kommandona.
func (p *Player) readLine() (line string, ok bool) {
	for p.scanner.Scan() {
		line := p.scanner.Text()
		if strings.HasPrefix(line, HelloPrefix) {
			p.structured.Store(true)
			p.SendFrame(Frame{Type: "hello", Text: "mudschool 1"})
			continue
		}
		return line, true
	}
	return "", false
}

// close stänger utkorgen och väntar tills allt i den har skickats.
// Efter close får ingen anropa Send.
func (p *Player) close() {
	close(p.outbox)
	<-p.done
}
