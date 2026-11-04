// Package mud contains all the game logic of your MUD server.
package mud

import (
	"bufio"
	"fmt"
	"net"
)

// Server is your MUD server. It has no fields yet, but it will grow:
// players, rooms, commands ...
type Server struct{}

// NewServer creates a new server.
func NewServer() *Server {
	return &Server{}
}

// Serve accepts connections from l, one at a time, until l is closed.
func (s *Server) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		s.handle(conn)
	}
}

// handle talks to one connected client until it disconnects.
func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	fmt.Fprint(conn, Welcome+"\r\n")

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Fprint(conn, scanner.Text()+"\r\n")
	}
}
