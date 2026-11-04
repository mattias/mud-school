// Paket mud innehåller all spellogik för din MUD-server.
package mud

import (
	"bufio"
	"fmt"
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
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		s.handle(conn)
	}
}

// handle pratar med en ansluten klient tills den kopplar ner.
func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	fmt.Fprint(conn, Welcome+"\r\n")

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Fprint(conn, scanner.Text()+"\r\n")
	}
}
