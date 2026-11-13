package mud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

// Copyover (lesson 35): the classic MUD trick from the 90s. The server starts
// a NEW version of itself and hands over the open connections – as file
// descriptors – together with the players' state. Players only see "The
// world shudders for a moment ..." and keep playing on new code.
//
// It needs file descriptors that can be passed to a new process: Linux,
// macOS and WSL. On Windows PrepareCopyover returns an error.

// CopyoverPlayer is what the new server needs to know about a player.
type CopyoverPlayer struct {
	Name       string   `json:"name"`
	Room       string   `json:"room"`
	X          int      `json:"x"`
	Y          int      `json:"y"`
	HP         int      `json:"hp"`
	Items      []string `json:"items"`
	Structured bool     `json:"structured"` // the client speaks JSON
	Tutorial   int      `json:"tutorial"`   // the tutorial step (lesson 37)
	Lang       string   `json:"lang"`       // the player's language (lesson 38)
	File       int      `json:"file"`       // index in the files passed along
}

// CopyoverState is everything that is handed over. files[0] is the listener;
// every player says which file is theirs.
type CopyoverState struct {
	Ticks   int64            `json:"ticks"`
	Players []CopyoverPlayer `json:"players"`
}

// fileOf duplicates the file descriptor behind a connection or listener.
func fileOf(x any) (*os.File, error) {
	f, ok := x.(interface{ File() (*os.File, error) })
	if !ok {
		return nil, fmt.Errorf("%T can't be handed over", x)
	}
	return f.File()
}

// PrepareCopyover collects the state and duplicates the file descriptors of
// the listener and every player's connection. Call it from outside the game
// loop. The caller owns the files (and must close them).
func (s *Server) PrepareCopyover() (*CopyoverState, []*os.File, error) {
	s.lmu.Lock()
	if len(s.listeners) == 0 {
		s.lmu.Unlock()
		return nil, nil, errors.New("no listener")
	}
	lf, err := fileOf(s.listeners[0])
	s.lmu.Unlock()
	if err != nil {
		return nil, nil, err
	}
	files := []*os.File{lf}
	state := &CopyoverState{}
	s.call(func() {
		state.Ticks = s.ticks
		for _, name := range s.playerNames() {
			p := s.players[name]
			f, ferr := fileOf(p.conn)
			if ferr != nil {
				err = ferr
				return
			}
			files = append(files, f)
			cp := CopyoverPlayer{Name: p.Name, Room: p.Room.ID, X: p.Pos.X, Y: p.Pos.Y, HP: p.HP,
				Structured: p.Structured(), Tutorial: p.tutorial, Lang: p.lang, File: len(files) - 1}
			for _, it := range p.Inventory {
				cp.Items = append(cp.Items, it.ID)
			}
			state.Players = append(state.Players, cp)
		}
	})
	if err != nil {
		for _, f := range files {
			f.Close()
		}
		return nil, nil, err
	}
	return state, files, nil
}

// HandOver stops this server WITHOUT saying goodbye or logging anyone out:
// the new server has the connections now. Closing our copies of the file
// descriptors doesn't close the connections – the new process holds copies too.
func (s *Server) HandOver(ctx context.Context) error {
	s.handingOver.Store(true)
	s.lmu.Lock()
	s.closing.Store(true)
	for _, l := range s.listeners {
		l.Close()
	}
	s.lmu.Unlock()
	s.call(func() {
		for p := range s.conns {
			p.conn.Close()
		}
	})
	done := make(chan struct{})
	go func() {
		s.handlers.Wait()
		close(done)
	}()
	var err error
	select {
	case <-done:
	case <-ctx.Done():
		err = ctx.Err()
	}
	s.start()
	s.stopOnce.Do(func() { close(s.stop) })
	<-s.loopDone
	return err
}

// Done is closed when the server has stopped (after Shutdown or HandOver).
func (s *Server) Done() <-chan struct{} { return s.loopDone }

// Adopt takes over the connections from a copyover: the players are put back
// in the world as they were, without logging in, and get MsgCopyoverDone.
// It returns the listener to Serve.
func (s *Server) Adopt(state *CopyoverState, files []*os.File) (net.Listener, error) {
	if len(files) == 0 {
		return nil, errors.New("no files")
	}
	l, err := net.FileListener(files[0])
	files[0].Close() // FileListener made its own copy
	if err != nil {
		return nil, err
	}
	s.start()
	s.call(func() { s.ticks = state.Ticks }) // the loop runs now: it owns s.ticks
	for _, cp := range state.Players {
		if cp.File <= 0 || cp.File >= len(files) {
			continue
		}
		conn, err := net.FileConn(files[cp.File])
		files[cp.File].Close()
		if err != nil {
			log.Printf("copyover: %s: %v", cp.Name, err)
			continue
		}
		s.adopt(conn, cp)
	}
	return l, nil
}

// adopt brings one player back.
func (s *Server) adopt(conn net.Conn, cp CopyoverPlayer) {
	p := newPlayer(conn)
	p.Name = cp.Name
	p.structured.Store(cp.Structured)
	s.call(func() {
		s.conns[p] = true
		s.players[p.Name] = p
		p.Room = s.world.Room(cp.Room)
		if p.Room == nil {
			p.Room = s.world.Room(s.world.Start)
		}
		z := s.world.zoneOf(p.Room)
		p.Pos = Pos{cp.X, cp.Y}
		if !Walkable(z.At(p.Pos)) {
			p.Pos = z.freeNear(z.Center(), nil)
		}
		p.MaxHP = PlayerMaxHP
		p.HP = cp.HP
		p.tutorial = cp.Tutorial // lesson 37
		p.lang = cp.Lang         // lesson 38
		for _, id := range cp.Items {
			if it := s.takeItem(id); it != nil {
				p.Inventory = append(p.Inventory, it)
			}
		}
		p.SendKind(KindInfo, MsgCopyoverDone)
		s.look(p)
	})
	s.lmu.Lock()
	s.handlers.Add(1)
	s.lmu.Unlock()
	go func() {
		defer s.handlers.Done()
		defer conn.Close()
		defer p.close()
		defer s.call(func() { delete(s.conns, p) })
		s.play(p)
	}()
}

// cmdCopyover: "@copyover" – restart on new code without disconnecting anyone.
func cmdCopyover(s *Server, p *Player, cmd Command) {
	if s.onCopyover == nil {
		p.SendKind(KindError, fmt.Sprintf(MsgCopyoverFailed, "not available"))
		return
	}
	if !s.copyoverRunning.CompareAndSwap(false, true) {
		p.SendKind(KindError, fmt.Sprintf(MsgCopyoverFailed, "a copyover is already in progress"))
		return
	}
	name := p.Name
	go func() { // PrepareCopyover and starting a process must not run in the loop
		err := s.copyover()
		if err != nil {
			s.copyoverRunning.Store(false) // nothing was handed over: try again
			log.Printf("copyover: %v", err)
			s.do(func() {
				if p := s.players[name]; p != nil {
					p.SendKind(KindError, fmt.Sprintf(MsgCopyoverFailed, err))
				}
			})
		}
	}()
}

// copyover prepares, starts the new server and hands over.
func (s *Server) copyover() error {
	state, files, err := s.PrepareCopyover()
	if err != nil {
		return err
	}
	defer func() {
		for _, f := range files {
			f.Close()
		}
	}()
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	if err := s.onCopyover(data, files); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.HandOver(ctx)
}
