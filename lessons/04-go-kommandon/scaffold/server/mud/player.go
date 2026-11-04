package mud

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// Player is a connected player.
type Player struct {
	Name string // filled in from lesson 5

	conn    net.Conn
	scanner *bufio.Scanner
	quit    bool // set to true when the player wants to leave
}

func newPlayer(conn net.Conn) *Player {
	return &Player{conn: conn, scanner: bufio.NewScanner(conn)}
}

// Send sends text to the player. Text with several lines works too.
func (p *Player) Send(text string) {
	text = strings.ReplaceAll(text, "\n", "\r\n")
	fmt.Fprint(p.conn, text+"\r\n")
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
