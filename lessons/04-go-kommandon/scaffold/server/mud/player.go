package mud

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// Player är en ansluten spelare.
type Player struct {
	Name string // fylls i från lektion 5

	conn    net.Conn
	scanner *bufio.Scanner
	quit    bool // sätts till true när spelaren vill sluta
}

func newPlayer(conn net.Conn) *Player {
	return &Player{conn: conn, scanner: bufio.NewScanner(conn)}
}

// Send skickar text till spelaren. Text med flera rader fungerar också.
func (p *Player) Send(text string) {
	text = strings.ReplaceAll(text, "\n", "\r\n")
	fmt.Fprint(p.conn, text+"\r\n")
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
