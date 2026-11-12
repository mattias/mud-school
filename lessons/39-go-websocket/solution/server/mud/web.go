package mud

import (
	"context"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// The web gateway (lesson 39): browsers can't speak telnet, but they can
// open a WebSocket. The gateway serves the web client's files over HTTP and
// turns every WebSocket into a net.Conn – and hands it to Serve through a
// listener of its own. The rest of the server never knows the difference.

// ConnListener is a net.Listener whose connections are handed to it, one
// by one, with Offer – instead of being accepted from the network.
type ConnListener struct {
	conns  chan net.Conn
	closed chan struct{}
	once   sync.Once
	addr   net.Addr
}

// NewConnListener creates a ConnListener; Addr returns addr.
func NewConnListener(addr net.Addr) *ConnListener {
	return &ConnListener{conns: make(chan net.Conn), closed: make(chan struct{}), addr: addr}
}

// Offer hands a connection to whoever calls Accept. It waits until someone
// does, and returns false (without touching c) if the listener is closed.
func (l *ConnListener) Offer(c net.Conn) bool {
	select {
	case l.conns <- c:
		return true
	case <-l.closed:
		return false
	}
}

// Accept waits for the next offered connection. After Close it returns an
// error that wraps net.ErrClosed – just like a closed TCP listener.
func (l *ConnListener) Accept() (net.Conn, error) {
	select {
	case c := <-l.conns:
		return c, nil
	case <-l.closed:
		return nil, &net.OpError{Op: "accept", Net: "websocket", Addr: l.addr, Err: net.ErrClosed}
	}
}

// Close stops the listener. It is safe to call more than once.
func (l *ConnListener) Close() error {
	l.once.Do(func() { close(l.closed) })
	return nil
}

// Addr is the address the listener was created with.
func (l *ConnListener) Addr() net.Addr { return l.addr }

// webAddr is the Addr of a gateway's listener.
type webAddr struct{}

func (webAddr) Network() string { return "websocket" }
func (webAddr) String() string  { return "/ws" }

// Gateway serves the web client: its files on every path, and the game on /ws.
type Gateway struct {
	files    http.Handler
	listener *ConnListener
	// Origins are other hosts allowed to open a WebSocket, e.g.
	// "mud.example.com". The page's own host is always allowed.
	Origins []string
}

// NewGateway creates a gateway that serves the files in files.
func NewGateway(files fs.FS) *Gateway {
	return &Gateway{files: http.FileServerFS(files), listener: NewConnListener(webAddr{})}
}

// Listener is where the gateway's connections come out: give it to Serve.
func (g *Gateway) Listener() net.Listener { return g.listener }

// ServeHTTP makes Gateway an http.Handler.
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/ws" {
		g.files.ServeHTTP(w, r)
		return
	}
	// Accept checks that the page comes from the same host (or from one of
	// g.Origins): otherwise any web page could log its visitors into your MUD.
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: g.Origins})
	if err != nil {
		log.Printf("websocket from %s: %v", r.RemoteAddr, err)
		return // Accept has already answered with an error
	}
	// The connection outlives this request, so it gets its own context –
	// not r.Context(), which ends when ServeHTTP returns.
	var conn net.Conn = newDeadlineConn(websocket.NetConn(context.Background(), c, websocket.MessageText))
	if !g.listener.Offer(conn) {
		c.Close(websocket.StatusGoingAway, "the server is shutting down")
	}
}

// deadlineConn gives a WebSocket the read deadlines of a TCP connection.
// coder/websocket closes the whole connection when a read deadline passes –
// but the server uses deadlines to stop a *read* (Shutdown, the login time
// limit in lesson 43) and still wants to say goodbye. So a goroutine of its
// own reads, and Read waits for it or for the deadline, whichever is first:
// an expired deadline is os.ErrDeadlineExceeded and the connection stays open.
type deadlineConn struct {
	net.Conn                // the WebSocket: Write, Close, the addresses
	data      chan []byte   // what the reader goroutine has read
	closed    chan struct{} // closed by Close
	closeOnce sync.Once
	err       error  // why the reader stopped; set before data is closed
	rest      []byte // read but not yet returned

	mu       sync.Mutex
	deadline time.Time
	changed  chan struct{} // closed (and replaced) when the deadline changes
}

func newDeadlineConn(c net.Conn) *deadlineConn {
	d := &deadlineConn{Conn: c, data: make(chan []byte), closed: make(chan struct{}), changed: make(chan struct{})}
	go d.readLoop()
	return d
}

func (d *deadlineConn) readLoop() {
	for {
		buf := make([]byte, 4096)
		n, err := d.Conn.Read(buf)
		if n > 0 {
			select {
			case d.data <- buf[:n]:
			case <-d.closed:
				return
			}
		}
		if err != nil {
			d.err = err
			close(d.data)
			return
		}
	}
}

func (d *deadlineConn) Read(p []byte) (int, error) {
	for len(d.rest) == 0 {
		d.mu.Lock()
		deadline, changed := d.deadline, d.changed
		d.mu.Unlock()
		var timer *time.Timer
		var expired <-chan time.Time
		if !deadline.IsZero() {
			wait := time.Until(deadline)
			if wait <= 0 {
				return 0, os.ErrDeadlineExceeded
			}
			timer = time.NewTimer(wait)
			expired = timer.C
		}
		var b []byte
		ok, timedOut := true, false
		select {
		case b, ok = <-d.data:
		case <-expired:
			timedOut = true
		case <-changed: // a new deadline: look again
		}
		if timer != nil {
			timer.Stop()
		}
		switch {
		case timedOut:
			return 0, os.ErrDeadlineExceeded
		case !ok:
			return 0, d.err
		}
		d.rest = b
	}
	n := copy(p, d.rest)
	d.rest = d.rest[n:]
	return n, nil
}

func (d *deadlineConn) SetReadDeadline(t time.Time) error {
	d.mu.Lock()
	d.deadline = t
	close(d.changed)
	d.changed = make(chan struct{})
	d.mu.Unlock()
	return nil
}

func (d *deadlineConn) SetDeadline(t time.Time) error {
	d.SetReadDeadline(t)
	return d.Conn.SetWriteDeadline(t)
}

func (d *deadlineConn) Close() error {
	d.closeOnce.Do(func() { close(d.closed) })
	return d.Conn.Close()
}
