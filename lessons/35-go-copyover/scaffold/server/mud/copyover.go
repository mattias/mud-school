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
	// TODO(lesson 35): under s.lmu, fileOf(s.listeners[0]) – that is files[0].
	// Then in the game loop (s.call): the ticks, and for every player (in name
	// order) fileOf(p.conn) and a CopyoverPlayer whose File is its index.
	// On an error close the files you have.
	return nil, nil, errors.New("TODO: PrepareCopyover")
}

// HandOver stops this server WITHOUT saying goodbye or logging anyone out:
// the new server has the connections now. Closing our copies of the file
// descriptors doesn't close the connections – the new process holds copies too.
func (s *Server) HandOver(ctx context.Context) error {
	// TODO(lesson 35): like Shutdown – but set s.handingOver first, send no
	// goodbye, and close the connections (p.conn.Close()) instead of setting a
	// read deadline.
	_ = ctx
	return errors.New("TODO: HandOver")
}

// Done is closed when the server has stopped (after Shutdown or HandOver).
func (s *Server) Done() <-chan struct{} { return s.loopDone }

// Adopt takes over the connections from a copyover: the players are put back
// in the world as they were, without logging in, and get MsgCopyoverDone.
// It returns the listener to Serve.
func (s *Server) Adopt(state *CopyoverState, files []*os.File) (net.Listener, error) {
	// TODO(lesson 35): net.FileListener(files[0]); s.start(); s.ticks in the
	// game loop (s.call – the loop is running now); for every
	// player net.FileConn(files[cp.File]) and s.adopt(conn, cp). Close the
	// files afterwards (FileListener and FileConn make their own copies).
	_ = log.Printf
	return nil, errors.New("TODO: Adopt")
}

// adopt brings one player back.
func (s *Server) adopt(conn net.Conn, cp CopyoverPlayer) {
	// TODO(lesson 35): newPlayer(conn), name and structured; in the game loop:
	// conns, players, room (or the start room), a walkable position, HP, the
	// items (takeItem), MsgCopyoverDone and look. Then a handler goroutine
	// (handlers.Add under lmu) that runs s.play(p) and cleans up like handle.
	_ = net.Conn(nil)
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
