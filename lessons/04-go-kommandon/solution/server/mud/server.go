// Package mud contains all the game logic of your MUD server.
package mud

import (
	"log"
	"net"
)

// Server is your MUD server.
type Server struct {
	commands map[string]CommandFunc
}

// NewServer creates a new server.
func NewServer() *Server {
	s := &Server{commands: make(map[string]CommandFunc)}
	s.registerBuiltins()
	return s
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

	p := newPlayer(conn)
	p.Send(Welcome)
	for !p.quit {
		line, ok := p.readLine()
		if !ok {
			return
		}
		s.dispatch(p, line)
	}
}
