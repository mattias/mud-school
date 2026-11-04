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
	world    *World

	mu      sync.Mutex         // protects the fields below
	players map[string]*Player // logged-in players, keyed by name
}

// NewServer creates a new server with the default world.
func NewServer() *Server {
	return NewServerWithWorld(DefaultWorld())
}

// NewServerWithWorld creates a new server with a world of your choice.
func NewServerWithWorld(w *World) *Server {
	s := &Server{
		commands: make(map[string]CommandFunc),
		world:    w,
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
	defer p.close() // runs before conn.Close, so everything gets sent

	p.Send(Welcome)
	if !s.login(p) {
		return
	}
	log.Printf("%s logged in", p.Name)
	p.Sendf(MsgWelcomeName, p.Name)
	s.broadcast(p, fmt.Sprintf(MsgJoined, p.Name))
	cmdLook(s, p, Command{Verb: "look"})
	defer func() {
		s.removePlayer(p)
		s.broadcast(p, fmt.Sprintf(MsgLeft, p.Name))
	}()

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
	p.Room = s.world.Room(s.world.Start)
	return nil
}

func (s *Server) removePlayer(p *Player) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.players, p.Name)
}

// broadcast sends text to every logged-in player except except (which may be nil).
// The lock is held while queueing; that is quick because Send never blocks.
func (s *Server) broadcast(except *Player, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, other := range s.players {
		if other != except {
			other.Send(text)
		}
	}
}

// PlayerNames returns the names of all logged-in players, sorted.
func (s *Server) PlayerNames() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.playerNamesLocked()
}

func (s *Server) playerNamesLocked() []string {
	names := make([]string, 0, len(s.players))
	for name := range s.players {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
