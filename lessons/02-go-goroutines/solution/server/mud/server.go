// Package mud contains all the game logic of your MUD server.
package mud

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

// Server is your MUD server.
type Server struct{}

// NewServer creates a new server.
func NewServer() *Server {
	return &Server{}
}

// Serve accepts connections from l until l is closed. Each connection is
// handled in its own goroutine.
func (s *Server) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go s.handle(conn)
	}
}

// handle talks to one connected client until it disconnects.
func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	log.Printf("new connection from %s", conn.RemoteAddr())
	defer log.Printf("%s disconnected", conn.RemoteAddr())

	fmt.Fprint(conn, Welcome+"\r\n")

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Fprint(conn, scanner.Text()+"\r\n")
	}
}
