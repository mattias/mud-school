// Package mud contains all the game logic of your MUD server.
package mud

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ErrServerClosed is returned by Serve after Shutdown has been called.
var ErrServerClosed = errors.New("mud: server closed")

// DefaultTick is how often the game world advances if Options.Tick is zero.
const DefaultTick = 100 * time.Millisecond

// Options configures a server. The zero value is a server with the default
// world and tick rate.
type Options struct {
	World *World        // nil = DefaultWorld()
	Tick  time.Duration // 0 = DefaultTick
}

// Server is your MUD server.
//
// The game world is owned by ONE goroutine, the game loop (see run). Other
// goroutines never touch it directly: they send it work with s.do or s.call.
// That is the "actor" pattern – no mutex needed for the world.
type Server struct {
	commands map[string]CommandFunc
	world    *World
	tick     time.Duration

	events    chan func()   // work for the game loop
	stop      chan struct{} // closed to stop the game loop
	loopDone  chan struct{} // closed when the game loop has stopped
	startOnce sync.Once
	stopOnce  sync.Once
	closing   atomic.Bool
	handlers  sync.WaitGroup // one per connection

	lmu       sync.Mutex // protects listeners
	listeners []net.Listener

	// Owned by the game loop – only touch these inside s.do / s.call.
	players map[string]*Player // logged-in players, keyed by name
	conns   map[*Player]bool   // every connection, logged in or not
	ticks   int64              // how many ticks the world has run
}

// NewServer creates a new server with the default world.
func NewServer() *Server {
	return NewServerWithOptions(Options{})
}

// NewServerWithWorld creates a new server with a world of your choice.
func NewServerWithWorld(w *World) *Server {
	return NewServerWithOptions(Options{World: w})
}

// NewServerWithOptions creates a new server.
func NewServerWithOptions(o Options) *Server {
	if o.World == nil {
		o.World = DefaultWorld()
	}
	if o.Tick <= 0 {
		o.Tick = DefaultTick
	}
	s := &Server{
		commands: make(map[string]CommandFunc),
		world:    o.World,
		tick:     o.Tick,
		events:   make(chan func()),
		stop:     make(chan struct{}),
		loopDone: make(chan struct{}),
		players:  make(map[string]*Player),
		conns:    make(map[*Player]bool),
	}
	s.registerBuiltins()
	return s
}

// Serve accepts connections from l until l is closed or Shutdown is called.
// Each connection is handled in its own goroutine.
func (s *Server) Serve(l net.Listener) error {
	s.start()
	s.lmu.Lock()
	s.listeners = append(s.listeners, l)
	s.lmu.Unlock()
	if s.closing.Load() {
		l.Close()
		return ErrServerClosed
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			if s.closing.Load() {
				return ErrServerClosed
			}
			return err
		}
		// Check closing and Add under lmu – Shutdown sets closing under the
		// same lock, so no connection can slip in after it starts waiting.
		s.lmu.Lock()
		if s.closing.Load() {
			s.lmu.Unlock()
			conn.Close()
			return ErrServerClosed
		}
		s.handlers.Add(1)
		s.lmu.Unlock()
		go func() {
			defer s.handlers.Done()
			s.handle(conn)
		}()
	}
}

// handle talks to one connected client until it disconnects.
func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	log.Printf("new connection from %s", conn.RemoteAddr())
	defer log.Printf("%s disconnected", conn.RemoteAddr())

	p := newPlayer(conn)
	defer p.close() // runs before conn.Close, so everything gets sent

	// Refuse the connection if Shutdown has already said goodbye to everyone.
	registered := false
	s.call(func() {
		if !s.closing.Load() {
			s.conns[p] = true
			registered = true
		}
	})
	if !registered {
		return
	}
	defer s.call(func() { delete(s.conns, p) })

	p.Send(Welcome)
	if !s.login(p) {
		return
	}
	log.Printf("%s logged in", p.Name)
	p.Sendf(MsgWelcomeName, p.Name)
	s.call(func() {
		s.broadcast(p, KindMove, fmt.Sprintf(MsgJoined, p.Name))
		s.look(p)
	})
	defer s.call(func() {
		s.removePlayer(p)
		s.broadcast(p, KindMove, fmt.Sprintf(MsgLeft, p.Name))
	})

	for !p.quit {
		line, ok := p.readLine()
		if !ok {
			return
		}
		// call (not do): commands for one player run in order, and p.quit is
		// set before we look at it again.
		if !s.call(func() { s.dispatch(p, line) }) {
			return
		}
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
		var err error
		if !s.call(func() { err = s.addPlayer(p) }) {
			return false
		}
		if err != nil {
			p.Sendf(MsgInvalidName, err)
			continue
		}
		return true
	}
}

// The functions below run in the game loop.

func (s *Server) addPlayer(p *Player) error {
	if _, taken := s.players[p.Name]; taken {
		return fmt.Errorf("%w: %s", ErrNameTaken, p.Name)
	}
	s.players[p.Name] = p
	p.Room = s.world.Room(s.world.Start)
	return nil
}

func (s *Server) removePlayer(p *Player) {
	if s.players[p.Name] == p {
		delete(s.players, p.Name)
	}
	// What the player carried goes back to where it started, so that it
	// isn't lost from the world (lesson 20 gives it back on the next login).
	for _, it := range p.Inventory {
		if r := s.world.Room(it.Room); r != nil {
			r.Items = append(r.Items, it)
		}
	}
	p.Inventory = nil
}

// broadcast sends text of a kind to every logged-in player except except (which may be nil).
func (s *Server) broadcast(except *Player, kind, text string) {
	for _, other := range s.players {
		if other != except {
			other.SendKind(kind, text)
		}
	}
}

// playerNames returns the names of all logged-in players, sorted.
func (s *Server) playerNames() []string {
	names := make([]string, 0, len(s.players))
	for name := range s.players {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// PlayerNames returns the names of all logged-in players, sorted.
// Safe to call from any goroutine (but not from inside the game loop).
func (s *Server) PlayerNames() []string {
	var names []string
	s.call(func() { names = s.playerNames() })
	return names
}
