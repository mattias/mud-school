package mud

import (
	"context"
	"time"
)

// The game loop. Everything that reads or changes the game world runs in
// this ONE goroutine; other goroutines send it work with do and call.
//
// You need these fields in Server (see README):
//
//	events    chan func()   // work for the game loop
//	stop      chan struct{} // closed to stop the game loop
//	loopDone  chan struct{} // closed when the game loop has stopped
//	startOnce sync.Once
//	stopOnce  sync.Once

// start starts the game loop (once).
func (s *Server) start() {
	s.startOnce.Do(func() { go s.run() })
}

// run is the game loop.
func (s *Server) run() {
	defer close(s.loopDone)
	// TODO(lesson 14): a time.Ticker with s.tick, and a for/select that
	//  - runs every function that arrives on s.events,
	//  - counts a tick (s.ticks++) and calls s.onTick() on every tick,
	//  - returns when s.stop is closed.
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
	// TODO(lesson 14): read s.ticks inside the game loop (s.call).
	return 0
}

// Shutdown stops the server: no new connections, everyone gets MsgShutdown
// and is disconnected, and the game loop stops. If ctx ends first, Shutdown
// stops waiting and returns ctx.Err().
func (s *Server) Shutdown(ctx context.Context) error {
	// TODO(lesson 14), see the README:
	//  1. mark the server as closing and close every listener Serve uses,
	//  2. in the game loop: send MsgShutdown to every connection and make its
	//     readLine return (p.conn.SetReadDeadline(time.Now())),
	//  3. wait until every handle goroutine is done – or ctx is done,
	//  4. stop the game loop and wait for it.
	_ = time.Now
	return nil
}
