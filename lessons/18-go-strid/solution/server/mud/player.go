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

// outboxSize is how many messages may wait to be sent to a player before we
// consider the client too slow.
const outboxSize = 256

// Player is a connected player.
type Player struct {
	Name      string
	Room      *Room   // where the player is; owned by the game loop
	Inventory []*Item // what the player carries; owned by the game loop
	HP, MaxHP int     // health; owned by the game loop
	target    *NPC    // who the player is fighting (nil = nobody)

	conn    net.Conn
	scanner *bufio.Scanner
	quit    bool // set to true when the player wants to leave

	structured atomic.Bool   // true = the client wants JSON (see HelloPrefix)
	outbox     chan Frame    // everything to be sent to the player
	done       chan struct{} // closed when the writer goroutine has finished
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

// writeLoop is the ONLY goroutine that writes to p.conn. It also decides how
// each Frame is encoded: as text or as JSON.
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
			continue // the connection is dead; keep draining so nobody gets stuck
		}
	}
}

// SendFrame queues a Frame for the player. It never blocks: if the outbox is
// full the player is disconnected instead.
func (p *Player) SendFrame(f Frame) {
	select {
	case p.outbox <- f:
	default:
		log.Printf("%s is too slow – disconnecting", p.conn.RemoteAddr())
		p.conn.Close()
	}
}

// Send sends text to the player.
func (p *Player) Send(text string) {
	p.SendFrame(TextFrame(text))
}

// SendKind sends text of a certain kind (KindSay, KindError, ...).
func (p *Player) SendKind(kind, text string) {
	p.SendFrame(KindFrame(kind, text))
}

// Sendf is like Send but with fmt.Sprintf formatting.
func (p *Player) Sendf(format string, args ...any) {
	p.Send(fmt.Sprintf(format, args...))
}

// Structured reports whether the client wants JSON.
func (p *Player) Structured() bool {
	return p.structured.Load()
}

// readLine reads the next line from the player. ok is false if the player disconnected.
// Protocol lines (like "@hello") are handled here and never reach the commands.
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

// close closes the outbox and waits until everything in it has been sent.
// Nobody may call Send after close.
func (p *Player) close() {
	close(p.outbox)
	<-p.done
}
