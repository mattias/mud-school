// Paket mud innehåller all spellogik för din MUD-server.
package mud

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

// Welcome är det första servern säger till en ny spelare.
const Welcome = "Välkommen till MUD School!"

// Server är din MUD-server.
type Server struct{}

// NewServer skapar en ny server.
func NewServer() *Server {
	return &Server{}
}

// Serve tar emot anslutningar från l tills l stängs. Varje anslutning
// hanteras i en egen goroutine.
func (s *Server) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go s.handle(conn)
	}
}

// handle pratar med en ansluten klient tills den kopplar ner.
func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	log.Printf("ny anslutning från %s", conn.RemoteAddr())
	defer log.Printf("%s kopplade ner", conn.RemoteAddr())

	fmt.Fprint(conn, Welcome+"\r\n")

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Fprint(conn, scanner.Text()+"\r\n")
	}
}
