package mud

import (
	"context"
	"time"
)

// The game loop. Everything that reads or changes the game world runs in
// this ONE goroutine; other goroutines send it work with do and call.

// start starts the game loop (once).
func (s *Server) start() {
	s.startOnce.Do(func() { go s.run() })
}

// run is the game loop. Everything that reads or changes the world happens here.
func (s *Server) run() {
	defer close(s.loopDone)
	ticker := time.NewTicker(s.tick)
	defer ticker.Stop()
	for {
		select {
		case f := <-s.events:
			f()
		case <-ticker.C:
			s.ticks++
			s.onTick()
		case <-s.stop:
			return
		}
	}
}

// onTick runs once per tick in the game loop. Later lessons add creatures,
// combat and saving here.
func (s *Server) onTick() {}

// do queues f to run in the game loop and returns at once.
// It returns false if the game loop has stopped.
func (s *Server) do(f func()) bool {
	s.start()
	select {
	case s.events <- f:
		return true
	case <-s.loopDone:
		return false
	}
}

// call runs f in the game loop and waits until it has run.
// Never call it from inside the game loop – that would wait for itself.
func (s *Server) call(f func()) bool {
	done := make(chan struct{})
	if !s.do(func() { f(); close(done) }) {
		return false
	}
	select {
	case <-done:
		return true
	case <-s.loopDone:
		return false
	}
}

// Ticks returns how many ticks the world has run.
func (s *Server) Ticks() int64 {
	var t int64
	s.call(func() { t = s.ticks })
	return t
}

// Shutdown stops the server: no new connections, everyone gets a goodbye
// message and is disconnected, and the game loop stops. If ctx ends first,
// Shutdown stops waiting and returns ctx.Err().
func (s *Server) Shutdown(ctx context.Context) error {
	s.lmu.Lock()
	s.closing.Store(true)
	for _, l := range s.listeners {
		l.Close()
	}
	s.lmu.Unlock()

	s.call(func() {
		for p := range s.conns {
			p.Send(MsgShutdown)
			// Make the player's readLine return, so that handle cleans up.
			p.conn.SetReadDeadline(time.Now())
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
	s.start() // so that loopDone is always closed
	s.stopOnce.Do(func() { close(s.stop) })
	<-s.loopDone
	return err
}
