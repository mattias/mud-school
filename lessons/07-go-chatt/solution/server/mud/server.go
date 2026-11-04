// Paket mud innehåller all spellogik för din MUD-server.
package mud

import (
	"fmt"
	"log"
	"net"
	"sort"
	"strings"
	"sync"
)

// Welcome är det första servern säger till en ny spelare.
const Welcome = "Välkommen till MUD School!"

// Server är din MUD-server.
type Server struct {
	commands map[string]CommandFunc

	mu      sync.Mutex         // skyddar fälten nedanför
	players map[string]*Player // inloggade spelare, nyckel = namn
}

// NewServer skapar en ny server.
func NewServer() *Server {
	s := &Server{
		commands: make(map[string]CommandFunc),
		players:  make(map[string]*Player),
	}
	s.registerBuiltins()
	return s
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

	p := newPlayer(conn)
	defer p.close() // körs före conn.Close, så att allt hinner skickas

	p.Send(Welcome)
	if !s.login(p) {
		return
	}
	log.Printf("%s loggade in", p.Name)
	p.Sendf("Välkommen, %s!", p.Name)
	s.broadcast(p, p.Name+" har anslutit.")
	defer func() {
		s.removePlayer(p)
		s.broadcast(p, p.Name+" har lämnat.")
	}()

	for !p.quit {
		line, ok := p.readLine()
		if !ok {
			return
		}
		s.dispatch(p, line)
	}
}

// login frågar efter ett namn tills spelaren anger ett giltigt och ledigt.
// Returnerar false om spelaren kopplar ner under tiden.
func (s *Server) login(p *Player) bool {
	for {
		p.Send("Vad heter du?")
		line, ok := p.readLine()
		if !ok {
			return false
		}
		name := strings.TrimSpace(line)
		if err := ValidateName(name); err != nil {
			p.Sendf("Ogiltigt namn: %v.", err)
			continue
		}
		p.Name = NormalizeName(name)
		if err := s.addPlayer(p); err != nil {
			p.Sendf("Ogiltigt namn: %v.", err)
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

// broadcast skickar text till alla inloggade spelare utom except (som får vara nil).
// Låset hålls medan vi köar meddelandena; det går snabbt eftersom Send aldrig blockerar.
func (s *Server) broadcast(except *Player, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, other := range s.players {
		if other != except {
			other.Send(text)
		}
	}
}

// PlayerNames returnerar namnen på alla inloggade spelare, sorterade.
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
