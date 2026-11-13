// Package mud contains all the game logic of your MUD server.
package mud

import (
	"cmp"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"os"
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
	Rand  *rand.Rand    // nil = a random seed; tests pass a fixed seed
	// Respawn is how many ticks a defeated creature stays away (0 = DefaultRespawn).
	Respawn int
	// Store saves accounts and players (lesson 20). nil = nothing is saved
	// and nobody needs a password.
	Store *Store
	// WorldPath is the world file (lesson 31); "" = the world isn't from a file.
	WorldPath string
	// ReloadEvery is how often the world file is checked for changes
	// (0 = never; @reload still works).
	ReloadEvery time.Duration
	// Admins may use the @ commands (lesson 31).
	Admins []string
	// ScriptDir holds .lua scripts (lesson 32); "" = no scripts. They are
	// reloaded when they change if ReloadEvery > 0.
	ScriptDir string
	// PluginDir holds .wasm plugins (lesson 33); "" = no plugins.
	PluginDir string
	// OnCopyover starts the new server for @copyover (lesson 35): it gets the
	// state as JSON and the files (the listener first, then one per player).
	// nil = @copyover is not available.
	OnCopyover func(state []byte, files []*os.File) error

	// Protection (lesson 43). Zero means no limit.
	MaxPerIP      int           // connections from one IP address at a time
	CommandRate   float64       // commands per second a player may send ...
	CommandBurst  int           // ... with bursts of this many
	MaxLineLength int           // the longest line; 0 = DefaultMaxLine
	LoginTimeout  time.Duration // how long logging in may take
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
	rng      *rand.Rand // only used in the game loop
	respawn  int
	store    *Store // nil = no saving

	worldPath   string          // lesson 31
	reloadEvery time.Duration   // lesson 31
	admins      map[string]bool // normalized names

	scriptDir string   // lesson 32
	scripts   *Scripts // owned by the game loop

	pluginDir string   // lesson 33
	plugins   *Plugins // owned by the game loop

	onCopyover      func(state []byte, files []*os.File) error // lesson 35
	handingOver     atomic.Bool                                // true during a copyover
	copyoverRunning atomic.Bool                                // one @copyover at a time

	saveMu   sync.Mutex     // held while the world file is written and while the watcher looks at it
	saveSeq  int64          // numbers the snapshots, in the game loop (lesson 36)
	savedSeq int64          // the newest snapshot written; protected by saveMu
	ownWrite atomic.Value   // fileStamp of the world file we wrote ourselves (lesson 36)
	saves    sync.WaitGroup // world saves still being written (lesson 36)

	hooks map[string][]Hook // lesson 34; owned by the game loop

	maxPerIP     int // lesson 43
	commandRate  float64
	commandBurst int
	maxLine      int
	loginTimeout time.Duration

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
	if o.Respawn <= 0 {
		o.Respawn = DefaultRespawn
	}
	if o.Rand == nil {
		o.Rand = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	}
	s := &Server{
		commands: make(map[string]CommandFunc),
		world:    o.World,
		tick:     o.Tick,
		rng:      o.Rand,
		respawn:  o.Respawn,
		store:    o.Store,

		worldPath:   o.WorldPath,
		reloadEvery: o.ReloadEvery,
		admins:      normalizeAdmins(o.Admins),
		scriptDir:   o.ScriptDir,
		pluginDir:   o.PluginDir,
		onCopyover:  o.OnCopyover,

		maxPerIP:     o.MaxPerIP,
		commandRate:  o.CommandRate,
		commandBurst: max(o.CommandBurst, 1),
		maxLine:      cmp.Or(o.MaxLineLength, DefaultMaxLine),
		loginTimeout: o.LoginTimeout,
		events:       make(chan func()),
		stop:         make(chan struct{}),
		loopDone:     make(chan struct{}),
		players:      make(map[string]*Player),
		conns:        make(map[*Player]bool),
	}
	s.registerBuiltins()
	s.On(EventCommand, s.onTutorialCommand) // lesson 37
	if s.scriptDir != "" {
		s.loadScripts() // the loop isn't running yet, so this is safe here
	}
	if s.pluginDir != "" {
		s.loadPlugins()
	}
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
	defer p.close()                  // runs before conn.Close, so everything gets sent
	p.ip = hostOf(conn.RemoteAddr()) // lesson 43
	// Longer lines end the connection. (The limit is the larger of max and the
	// buffer's capacity – so the buffer may not start bigger than the limit.)
	p.scanner.Buffer(make([]byte, 0, min(1024, s.maxLine)), s.maxLine)
	metricConnections.Add(1)

	// Refuse the connection if Shutdown has already said goodbye to everyone
	// – or if this address already has too many (lesson 43).
	registered, tooMany := false, false
	s.call(func() {
		switch {
		case s.closing.Load():
		case s.maxPerIP > 0 && s.connsFrom(p.ip) >= s.maxPerIP:
			tooMany = true
		default:
			if s.loginTimeout > 0 {
				// Nobody may hold a connection without logging in (lesson 43). Set
				// here, in the loop, before Shutdown can see p: otherwise this could
				// overwrite the deadline Shutdown sets to make the read stop.
				conn.SetReadDeadline(time.Now().Add(s.loginTimeout))
			}
			s.conns[p] = true
			registered = true
		}
	})
	if tooMany {
		metricRefused.Add(1)
		p.SendKind(KindError, MsgTooManyConnections)
		return
	}
	if !registered {
		return
	}
	defer s.call(func() { delete(s.conns, p) })

	p.Send(Welcome)
	saved, ok := s.login(p)
	if !ok {
		if errors.Is(p.scanner.Err(), os.ErrDeadlineExceeded) && !s.closing.Load() {
			p.SendKind(KindError, MsgLoginTimeout)
		}
		return
	}
	log.Printf("%s logged in", p.Name)
	p.Sendf(MsgWelcomeName, p.Name)
	s.call(func() {
		// Logged in: no more time limit – unless Shutdown has just set one.
		if !s.closing.Load() {
			conn.SetReadDeadline(time.Time{})
		}
		if saved != nil {
			s.restore(p, *saved)
		} else {
			s.startTutorial(p) // a new player (lesson 37)
		}
		s.broadcast(p, KindMove, fmt.Sprintf(MsgJoined, p.Name))
		defer s.syncAll()
		s.look(p)
		s.showTutorialStep(p)
	})
	s.play(p)
}

// play runs a logged-in player's commands until they leave, and cleans up
// afterwards. Players handed over by a copyover (lesson 35) start here.
func (s *Server) play(p *Player) {
	defer func() {
		if s.handingOver.Load() {
			return // a copyover: the new server has the player now (lesson 35)
		}
		// 1. A snapshot in the game loop, 2. write it to the database outside
		// the loop (disk is slow), 3. only then free the name – so that a quick
		// reconnect can't load the old save before the new one is written.
		var snap SavedPlayer
		s.call(func() {
			p.target = nil // stop fighting, so nothing changes after the snapshot
			snap = p.snapshot()
		})
		if s.store != nil && snap.Name != "" {
			if err := s.store.SavePlayer(snap); err != nil {
				log.Printf("cannot save %s: %v", p.Name, err)
			}
		}
		s.call(func() {
			s.removePlayer(p)
			s.broadcast(p, KindMove, fmt.Sprintf(MsgLeft, p.Name))
			s.syncAll()
		})
	}()

	if s.commandRate > 0 { // lesson 43
		p.limiter = NewRateLimiter(s.commandRate, s.commandBurst)
	}
	for !p.quit {
		line, ok := p.readLine()
		if !ok {
			return
		}
		if ok, quit := s.allowCommand(p); quit {
			return
		} else if !ok {
			continue // dropped
		}
		// call (not do): commands for one player run in order, and p.quit is
		// set before we look at it again.
		if !s.call(func() { s.dispatch(p, line); s.syncAll() }) {
			return
		}
	}
}

// login asks for a name (and, with a store, a password) until the player
// gives a valid, free one. saved is what was stored about the player last
// time (nil = nothing). ok is false if the player disconnects meanwhile.
func (s *Server) login(p *Player) (saved *SavedPlayer, ok bool) {
	for {
		saved = nil
		p.Send(MsgAskName)
		line, ok := p.readLine()
		if !ok {
			return nil, false
		}
		name := strings.TrimSpace(line)
		if err := ValidateName(name); err != nil {
			p.Sendf(MsgInvalidName, err)
			continue
		}
		p.Name = NormalizeName(name)
		if s.store != nil {
			ok, retry := s.authenticate(p)
			if !ok {
				return nil, false
			}
			if retry {
				continue
			}
			sp, found, err := s.store.LoadPlayer(p.Name)
			if err != nil {
				log.Printf("cannot load %s: %v", p.Name, err)
				p.SendKind(KindError, MsgStoreError)
				return nil, false
			}
			if found {
				saved = &sp
			}
		}
		var err error
		if !s.call(func() { err = s.addPlayer(p) }) {
			return nil, false
		}
		if err != nil {
			p.Sendf(MsgInvalidName, err)
			continue
		}
		return saved, true
	}
}

// The functions below run in the game loop.

func (s *Server) addPlayer(p *Player) error {
	if _, taken := s.players[p.Name]; taken {
		return fmt.Errorf("%w: %s", ErrNameTaken, p.Name)
	}
	s.players[p.Name] = p
	metricPlayers.Add(1) // lesson 43
	p.Room = s.world.Room(s.world.Start)
	p.Pos = s.world.zoneOf(p.Room).freeNear(s.world.zoneOf(p.Room).Center(), nil)
	p.MaxHP = PlayerMaxHP
	p.HP = p.MaxHP
	return nil
}

func (s *Server) removePlayer(p *Player) {
	if s.players[p.Name] == p {
		delete(s.players, p.Name)
		metricPlayers.Add(-1) // lesson 43
	}
	// What the player carried goes back to where it started, so that it
	// isn't lost from the world (lesson 20 gives it back on the next login).
	for _, it := range p.Inventory {
		if r := s.world.Room(it.Room); r != nil {
			s.world.placeItem(r, it)
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
