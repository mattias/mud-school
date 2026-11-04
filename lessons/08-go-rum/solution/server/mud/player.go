package mud

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

// outboxSize is how many messages may wait to be sent to a player before we
// consider the client too slow.
const outboxSize = 256

// Player is a connected player.
type Player struct {
	Name string
	Room *Room // where the player is; protected by Server.mu

	conn    net.Conn
	scanner *bufio.Scanner
	quit    bool // set to true when the player wants to leave

	outbox chan string   // everything to be sent to the player
	done   chan struct{} // closed when the writer goroutine has finished
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

// writeLoop is the ONLY goroutine that writes to p.conn.
func (p *Player) writeLoop() {
	defer close(p.done)
	for msg := range p.outbox {
		if _, err := io.WriteString(p.conn, msg); err != nil {
			// The connection is dead. Keep draining so nobody gets stuck.
			continue
		}
	}
}

// Send queues text for the player. Send never blocks: if the outbox is full
// the player is disconnected instead.
func (p *Player) Send(text string) {
	msg := strings.ReplaceAll(text, "\n", "\r\n") + "\r\n"
	select {
	case p.outbox <- msg:
	default:
		log.Printf("%s is too slow – disconnecting", p.conn.RemoteAddr())
		p.conn.Close()
	}
}

// Sendf is like Send but with fmt.Sprintf formatting.
func (p *Player) Sendf(format string, args ...any) {
	p.Send(fmt.Sprintf(format, args...))
}

// readLine reads the next line from the player. ok is false if the player disconnected.
func (p *Player) readLine() (line string, ok bool) {
	if !p.scanner.Scan() {
		return "", false
	}
	return p.scanner.Text(), true
}

// close closes the outbox and waits until everything in it has been sent.
// Nobody may call Send after close.
func (p *Player) close() {
	close(p.outbox)
	<-p.done
}
