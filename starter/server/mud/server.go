// Paket mud innehåller all spellogik för din MUD-server.
package mud

import (
	"net"
)

// Welcome är det första servern säger till en ny spelare.
const Welcome = "Välkommen till MUD School!"

// Server är din MUD-server. Just nu har den inga fält, men den kommer att
// växa: spelare, rum, kommandon ...
type Server struct{}

// NewServer skapar en ny server.
func NewServer() *Server {
	return &Server{}
}

// Serve tar emot anslutningar från l, en i taget, tills l stängs.
func (s *Server) Serve(l net.Listener) error {
	// TODO(lektion 1): ta emot anslutningar i en loop med l.Accept().
	// Om Accept returnerar ett fel ska Serve returnera felet.
	// Varje ny anslutning ska hanteras av s.handle(conn).
	return nil
}

// handle pratar med en ansluten klient tills den kopplar ner.
func (s *Server) handle(conn net.Conn) {
	// TODO(lektion 1):
	//  1. Se till att conn stängs när funktionen är klar (tips: defer).
	//  2. Skicka Welcome följt av "\r\n".
	//  3. Läs rader från klienten (tips: bufio.NewScanner) och skicka
	//     tillbaka varje rad (utan radbrytningen) följt av "\r\n".
}
