package mud

import (
	"context"
	"io/fs"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

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
	// TrustProxy: the player's address is the last one in X-Forwarded-For
	// (lesson 42). Only behind a proxy like Caddy, which sets that header –
	// otherwise anyone can claim to be anyone.
	TrustProxy bool
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
	var conn net.Conn = websocket.NetConn(context.Background(), c, websocket.MessageText)
	if g.TrustProxy { // lesson 42
		if ip := forwardedFor(r); ip != nil {
			conn = addrConn{Conn: conn, remote: &net.TCPAddr{IP: ip}}
		}
	}
	if !g.listener.Offer(conn) {
		c.Close(websocket.StatusGoingAway, "the server is shutting down")
	}
}

// forwardedFor is the client's IP according to the proxy: the last address
// in X-Forwarded-For (the one the proxy itself added). nil if there is none.
func forwardedFor(r *http.Request) net.IP {
	all := r.Header.Values("X-Forwarded-For")
	if len(all) == 0 {
		return nil
	}
	parts := strings.Split(all[len(all)-1], ",")
	return net.ParseIP(strings.TrimSpace(parts[len(parts)-1]))
}

// addrConn is a net.Conn with another RemoteAddr. Embedding the interface
// gives it every other method of the connection inside.
type addrConn struct {
	net.Conn
	remote net.Addr
}

func (c addrConn) RemoteAddr() net.Addr { return c.remote }
