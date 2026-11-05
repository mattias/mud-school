// Package mud contains all the game logic of your MUD server.
package mud

import (
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
// l is already listening: main.go (and the tests) create it with net.Listen.
func (s *Server) Serve(l net.Listener) error {
	// TODO(lesson 1): accept connections in a loop with l.Accept().
	// Use the l you were given – don't call net.Listen here.
	// If Accept returns an error, Serve must return that error.
	// Each new connection is handled by s.handle(conn).
	return nil
}

// handle talks to one connected client until it disconnects.
func (s *Server) handle(conn net.Conn) {
	// TODO(lesson 1):
	//  1. Make sure conn is closed when the function returns (hint: defer).
	//  2. Send Welcome (see messages.go) followed by "\r\n".
	//  3. Read lines from the client (hint: bufio.NewScanner) and send each
	//     line back (without its line break) followed by "\r\n".
}
