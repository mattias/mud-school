// Package mud contains all the game logic of your MUD server.
package mud

import (
	"fmt"
	"log"
	"net"
	"sort"
	"strings"
	"sync"
)

// Server is your MUD server.
type Server struct {
	commands map[string]CommandFunc

	mu      sync.Mutex         // protects the fields below
	players map[string]*Player // logged-in players, keyed by name
}

// NewServer creates a new server.
func NewServer() *Server {
	s := &Server{
		commands: make(map[string]CommandFunc),
		players:  make(map[string]*Player),
	}
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
	if !s.login(p) {
		return
	}
	defer s.removePlayer(p)
	log.Printf("%s logged in", p.Name)
	p.Sendf(MsgWelcomeName, p.Name)

	for !p.quit {
		line, ok := p.readLine()
		if !ok {
			return
		}
		s.dispatch(p, line)
	}
}

// login asks for a name until the player gives a valid, free one.
// Returns false if the player disconnects meanwhile.
func (s *Server) login(p *Player) bool {
	for {
		p.Send(MsgAskName)
		line, ok := p.readLine()
		if !ok {
			return false
		}
		name := strings.TrimSpace(line)
		if err := ValidateName(name); err != nil {
			p.Sendf(MsgInvalidName, err)
			continue
		}
		p.Name = NormalizeName(name)
		if err := s.addPlayer(p); err != nil {
			p.Sendf(MsgInvalidName, err)
			continue
		}
		return true
	}
}

func (s *Server) addPlayer(p *Player) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, taken := s.players[p.Name]; taken {
		return fmt.Errorf("%w: %s", ErrNameTaken, p.Name)
	}
	s.players[p.Name] = p
	return nil
}

func (s *Server) removePlayer(p *Player) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.players, p.Name)
}

// PlayerNames returns the names of all logged-in players, sorted.
func (s *Server) PlayerNames() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	names := make([]string, 0, len(s.players))
	for name := range s.players {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
